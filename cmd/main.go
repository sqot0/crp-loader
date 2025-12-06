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

	packs := map[string]string{
		"1": "https://pub-7ac6523b994b44f9b233ee0cbd3afccc.r2.dev/main.zip",
		"2": "https://pub-7ac6523b994b44f9b233ee0cbd3afccc.r2.dev/war.zip",
	}

	for {
		internal.ClearScreen()

		fmt.Println("==============================")
		fmt.Println("  Minecraft Pack Installer")
		fmt.Println("==============================")
		fmt.Println("")
		fmt.Println("Выберите сборку:")
		fmt.Println("")
		fmt.Println("1) Основная сборка")
		fmt.Println("2) Военная сборка")
		fmt.Println("")
		fmt.Print("Введите номер: ")

		modpackChoice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка ввода:", err)
			continue
		}
		modpackChoice = strings.TrimSpace(modpackChoice)

		downloadURL, exists := packs[modpackChoice]
		if !exists {
			fmt.Println("Неверный ввод.")
			fmt.Println("\nНажмите Enter чтобы продолжить...")
			if _, err := reader.ReadString('\n'); err != nil {
			}
			continue
		}
		internal.ClearScreen()

		if err := install(downloadURL); err != nil {
			fmt.Println("Ошибка установки:", err)
		}
		break
	}
}

func install(downloadURL string) error {
	reader := bufio.NewReader(os.Stdin)
	mcDir := "./"

	if err := os.RemoveAll(path.Join(mcDir, "mods")); err != nil {
		// ignore
	}

	tmp := "pack.zip"
	fmt.Print("Скачиваю pack.zip... ")
	if err := internal.DownloadFile(downloadURL, tmp); err != nil {
		fmt.Println("Ошибка:", err)
		return err
	}
	fmt.Println("OK")
	defer func() {
		_ = os.Remove(tmp)
	}()

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
	fmt.Println("OK")

	fmt.Println("Готово! Сборка установлена.")

	if _, readerErr := reader.ReadString('\n'); readerErr != nil {
		// ignore
	}

	return nil
}
