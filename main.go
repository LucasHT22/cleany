package main

import (
	"fmt"
	"os"

	"cleany/scanner"
	"cleany/ui"
)

func main() {
	root := `C:\`
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	fmt.Println("\033[2J\033[H")
	ui.PrintBanner()
	fmt.Printf("  Scanning %s ... \n\n", root)

	entries, total := scanner.Scan(root)
	ui.Run(entries, total, root)
}