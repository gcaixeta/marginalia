package main

import (
	"fmt"
	"os"
)

func remove(textType string) {
	fmt.Print("Removing", textType)
}

func list(textType string) {
	fmt.Print("Listing", textType)
}

func add(textType string) {
	fmt.Println("Adding", textType)
}

func main() {
	fmt.Println("vindo ao marginalia!")

	if len(os.Args) < 3 {
		fmt.Println("Exemplo de uso: margi [acao] [tipo-de-texto]")
		return
	}

	action := os.Args[1]
	textType := os.Args[2]

	switch action {
	case "add":
		add(textType)
	case "list":
		list(textType)
	case "remove":
		remove(textType)
	}
}
