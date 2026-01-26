package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/sqot0/crp-loader/internal"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	packs := []string{
		"https://pub-7ac6523b994b44f9b233ee0cbd3afccc.r2.dev/main.zip",
		"https://pub-7ac6523b994b44f9b233ee0cbd3afccc.r2.dev/war.zip",
	}

	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка определения пути:", err)
		return
	}
	mcDir := path.Dir(executable)

	localPacks := []string{}

	for _, packURL := range packs {
		localFile := path.Join(mcDir, path.Base(packURL))
		if _, err := os.Stat(localFile); err == nil {
			localPacks = append(localPacks, localFile)
		}
	}

	for {
		internal.ClearScreen()

		fmt.Println("   ____ ____  ____    _     ___    _    ____  _____ ____")
		fmt.Println("  / ___|  _ \\|  _ \\  | |   / _ \\  / \\  |  _ \\| ____|  _ \\")
		fmt.Println(" | |   | |_) | |_) | | |  | | | |/ _ \\ | | | |  _| | |_) |")
		fmt.Println(" | |___|  _ <|  __/  | |__| |_| / ___ \\| |_| | |___|  _ <")
		fmt.Println("  \\____|_| \\_\\_|     |_____\\___/_/   \\_\\____/|_____|_| \\_\\")
		fmt.Print("\n\n")
		fmt.Println("Выберите сборку:")
		fmt.Println("")
		fmt.Println("1) Основная сборка")
		fmt.Println("2) Военная сборка")
		for i, name := range localPacks {
			fmt.Printf("%d) %s (Локально)\n", len(packs)+i+1, name)
		}
		fmt.Println("")
		fmt.Print("Введите номер: ")

		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка ввода:", err)
			continue
		}
		modpackChoice, err := strconv.Atoi(strings.TrimSpace(userInput))
		if err != nil || modpackChoice < 1 || modpackChoice > (len(packs)+len(localPacks)) {
			fmt.Println("Неверный ввод.")
			fmt.Println("\nНажмите Enter чтобы продолжить...")
			if _, err := reader.ReadString('\n'); err != nil {
			}
			continue
		}
		modpackChoice -= 1 // Convert to 0-based index

		var downloadURL string
		if modpackChoice >= len(packs) {
			downloadURL = localPacks[modpackChoice-len(packs)]
		} else {
			downloadURL = packs[modpackChoice]
		}

		internal.ClearScreen()

		if err := install(downloadURL); err != nil {
			fmt.Println("Ошибка установки:", err)

			fmt.Println("Попробуйте скачать сборку вручную и поместить его в папку с загрузчиком (майнкрафта).")
			fmt.Println("Ссылка для скачивания:", downloadURL)
		}
		break
	}

	if _, readerErr := reader.ReadString('\n'); readerErr != nil {
		// ignore
	}
}

func install(downloadURL string) error {
	reader := bufio.NewReader(os.Stdin)

	executable, err := os.Executable()
	if err != nil {
		return err
	}
	mcDir := path.Dir(executable)

	if err := os.RemoveAll(path.Join(mcDir, "mods")); err != nil {
		// ignore
	}

	tmp := "pack.zip"
	if strings.HasPrefix(downloadURL, "http") {
		fmt.Print("Скачиваю pack.zip... ")
		if err := internal.DownloadFile(downloadURL, tmp); err != nil {
			return err
		}
		defer os.Remove(tmp)
	} else {
		tmp = downloadURL
		fmt.Println("Использую локальный файл:", tmp)
	}

	// Inspect zip for optional groups
	groups, err := internal.InspectOptionalGroups(tmp)
	if err != nil {
		fmt.Println("Не удалось просканировать zip:", err)
		return err
	}

	selected := []string{}
	if len(groups) > 0 {
		fmt.Println("\nНайдены опциональные наборы (optional):")
		sort.Strings(groups)
		for i, g := range groups {
			fmt.Printf("%d) %s\n", i+1, g)
		}
		fmt.Println("Введите номера наборов через запятую (например: 1,3) или 'all' для всех, пустая строка для пропуска:")
		fmt.Print("Выбор: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			choice = ""
		} else {
			choice = strings.TrimSpace(choice)
		}
		if choice == "all" || choice == "ALL" {
			selected = append(selected, groups...)
		} else if choice != "" {
			for _, part := range strings.Split(choice, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.Atoi(part)
				if err != nil {
					fmt.Println("Неверный номер:", part)
					continue
				}
				if n >= 1 && n <= len(groups) {
					selected = append(selected, groups[n-1])
				} else {
					fmt.Println("Номер вне диапазона:", n)
				}
			}
		}
	} else {
		fmt.Println("Опциональные наборы не найдены в архиве.")
	}

	fmt.Print("Распаковываю выбранные файлы... ")
	if err := internal.ExtractSelectedFromZip(tmp, mcDir, selected); err != nil {
		fmt.Println("Ошибка:", err)
		return err
	}

	fmt.Println("Готово! Сборка установлена.")

	return nil
}
