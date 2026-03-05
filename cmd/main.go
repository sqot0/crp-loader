package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/sqot0/crp-loader/internal"
	"github.com/sqot0/crp-loader/internal/manifest"
	"github.com/sqot0/crp-loader/internal/prompt"
	"github.com/sqot0/crp-loader/internal/terminal"
)

const baseURL = "https://pub-7ac6523b994b44f9b233ee0cbd3afccc.r2.dev/"

func main() {
	reader := bufio.NewReader(os.Stdin)

	mf, err := manifest.GetManifest(baseURL)
	if err != nil {
		fmt.Println("Ошибка получения манифеста:", err)
		return
	}

	terminal.ClearScreen()

	fmt.Println("   ____ ____  ____    _     ___    _    ____  _____ ____")
	fmt.Println("  / ___|  _ \\|  _ \\  | |   / _ \\  / \\  |  _ \\| ____|  _ \\")
	fmt.Println(" | |   | |_) | |_) | | |  | | | |/ _ \\ | | | |  _| | |_) |")
	fmt.Println(" | |___|  _ <|  __/  | |__| |_| / ___ \\| |_| | |___|  _ <")
	fmt.Println("  \\____|_| \\_\\_|     |_____\\___/_/   \\_\\____/|_____|_| \\_\\")
	fmt.Print("\n\n")

	var modpacks []manifest.ModpackInfo
	for _, mp := range mf.Modpacks {
		modpacks = append(modpacks, mp)
	}
	sort.Slice(modpacks, func(i, j int) bool {
		return modpacks[i].Priority < modpacks[j].Priority
	})
	modpackNames := make([]string, 0, len(modpacks))
	for _, mp := range modpacks {
		modpackNames = append(modpackNames, mp.Name)
	}

	selectedName, err := prompt.Select(modpackNames)
	if err != nil {
		fmt.Println("Ошибка выбора:", err)
		return
	}

	terminal.ClearScreen()

	if selectedName == "" {
		fmt.Println("Ошибка выбора: Вы не выбрали сборку")
		return
	}

	chosenPackKey, chosenPack := mf.GetModpackByName(selectedName)
	if chosenPack == nil {
		fmt.Println("Ошибка выбора: Сборка не найдена в манифесте")
		return
	}

	selectedOptionals := make([]string, 0)

	if len(chosenPack.Optionals) > 0 {
		displayStrings := make([]string, len(chosenPack.Optionals))
		for i, optKey := range chosenPack.Optionals {
			opt := mf.Optionals[optKey]
			displayStrings[i] = opt.Name + "\n" + opt.Description
		}
		selectedIndices, err := prompt.Multiselect(displayStrings)
		if err != nil {
			fmt.Println("Ошибка выбора опциональных наборов:", err)
			return
		}

		for _, idx := range selectedIndices {
			selectedOptionals = append(selectedOptionals, chosenPack.Optionals[idx])
		}

		terminal.ClearScreen()
	}

	type DownloadTask struct {
		ID   string
		URL  string
		SHA1 string
		Name string
	}

	tasks := []DownloadTask{}
	tasks = append(tasks, DownloadTask{
		ID:   chosenPackKey,
		URL:  baseURL + chosenPackKey + ".zip",
		SHA1: chosenPack.SHA1,
		Name: "Сборка: " + chosenPack.Name,
	})
	for _, optKey := range selectedOptionals {
		opt := mf.Optionals[optKey]
		tasks = append(tasks, DownloadTask{
			ID:   optKey,
			URL:  baseURL + "optionals/" + optKey + ".zip",
			SHA1: opt.SHA1,
			Name: "Опционально: " + opt.Name,
		})
	}

	progressChan := make(chan internal.ProgressUpdate, len(tasks)*5)
	installTasks := make([]*prompt.InstallTask, 0, len(tasks))
	for _, task := range tasks {
		installTasks = append(installTasks, &prompt.InstallTask{
			ID:     task.ID,
			Name:   task.Name,
			Status: internal.StatusWaiting,
		})
	}

	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка определения пути к исполняемому файлу:", err)
		return
	}
	mcDir := path.Dir(executable)

	if err := os.RemoveAll(path.Join(mcDir, "mods")); err != nil {
		// ignore
	}

	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t DownloadTask) {
			defer wg.Done()
			internal.Install(t.ID, t.URL, t.SHA1, progressChan)
		}(task)
	}

	go func() {
		wg.Wait()
		close(progressChan)
	}()

	if err := prompt.RunProgress(installTasks, progressChan); err != nil {
		fmt.Println("Ошибка UI:", err)
		return
	}

	fmt.Println("\nНажмите Enter для выхода")
	_, _ = reader.ReadString('\n')
}
