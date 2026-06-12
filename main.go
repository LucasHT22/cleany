package main

import (
	"flag"
	"fmt"
	"os"

	"cleany/scanner"
	"cleany/ui"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Simulates file detection without actually erasing them")
	flag.Parse()

	root := `C:\\`
	if len(flag.Args()) > 0 {
		root = flag.Args()[0]
	}

	fmt.Println("\033[2J\033[H")
	ui.PrintBanner()
	fmt.Printf("  Scanning %s ... \n\n", root)

	entries, total := scanner.Scan(root)
	ui.Run(entries, total, root, dryRun)
}