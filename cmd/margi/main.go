package main

import (
	"fmt"
	"os"
)

func newFile(textGenre string) {
	fmt.Println("new file of type", textGenre)
}

func listFiles(textGenre string) {
	fmt.Println("list files of type", textGenre)
}

func removeFile(textGenre string) {
	fmt.Println("remove file of type", textGenre)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage example: margi new note")
		return
	}

	action := os.Args[1]
	textType := os.Args[2]

	switch action {
	case "new":
		newFile(textType)
	case "list":
		listFiles(textType)
	case "remove":
		removeFile(textType)
	}
	return
}
