package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {

	username := os.Getenv("USERNAME")

	if username != "" {
		fmt.Printf("Пользователь: %s\n", username)
	} else {
		fmt.Println("Пользователь: (переменная окружения не задана)")
	}

	args := os.Args[1:]
	if len(args) > 0 {
		fmt.Println("Аргументы CLI:")
		for i, arg := range args {
			fmt.Printf("  [%d] %s\n", i, arg)
		}
	} else {
		fmt.Println("Аргументы CLI: (не переданы)")
	}

	fmt.Printf("Версия Go: %s\n", runtime.Version())
}
