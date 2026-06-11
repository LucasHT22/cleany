package ui

import (
	"fmt"
	"os"
	"strings"

	"cleany/scanner"
)

func viewNodeModules(root string) {
	fmt.Println(dim + "  Hunting node_modules..." + reset)

	results := scanner.ScanNodeModules(root)

	clearScreen()
	PrintBanner()
	fmt.Println(bold + cyan + "  NODE_MODULES HUNTER" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	if len(results) == 0 {
		fmt.Println(green + "  No node_modules found." + reset)
		fmt.Println()
		prompt(dim + "  Press Enter to go back..." + reset)
		return
	}

	max := results[0].SizeMB
	if max == 0 {
		max = 1
	}

	for i, r := range results {
		filled := int(30 * float64(r.SizeMB) / float64(max))
		if filled < 1 {
			filled = 1
		}
		bar := barColor(i) + strings.Repeat("█", filled) + dim + strings.Repeat("░", 30-filled) + reset
		fmt.Printf("  %2d. %-20s %s  %s%d MB%s\n		%s%s%s\n\n", i+1, truncate(r.Project, 20), bar, bold+barColor(i), r.SizeMB, reset, dim, truncate(r.Path, 60), reset)
	}

	var totalMB int64
	for _, r := range results {
		totalMB += r.SizeMB
	}
	fmt.Printf("  %sFound %d node_modules - %s%d MB total%s\n\n", dim, len(results), bold+white, totalMB, reset)
	fmt.Println(dim + "  Enter a number to delete, 'all' to delete everything, or 'q' to go back!" + reset)
	fmt.Println(dim + "  Safe to delete" + reset)
	fmt.Println()

	for {
		choice := prompt(bold + "  > " + reset)

		if choice == "q" || choice == "Q" {
			return
		}

		if strings.ToLower(choice) == "all" {
			confirm := prompt(fmt.Sprintf(red+"  Delete ALL %d node_modules folders (~%d MB)? [y/N] "+reset, len(results), totalMB))
			if strings.ToLower(confirm) == "y" {
				deleted := 0
				for _, r := range results {
					if err := os.RemoveAll(r.Path); err != nil {
						fmt.Printf(red+"  Failed: %s - %s\n"+reset, r.Path, err.Error())
					} else {
						fmt.Printf(green+"  Deleted: %s\n"+reset, r.Path)
						deleted++
					}
				}
				fmt.Printf(green+"\n  Done. Deleted %d/%d folders!\n"+reset, deleted, len(results))

				prompt(dim + "  Press Enter to go back..." + reset)
				return
			}
			continue
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx > len(results) {
			fmt.Println(red + "  Invalid choice." + reset)
			continue
		}

		target := results[idx-1]
		confirm := prompt(fmt.Sprintf(red+"  Delete node_modules in '%d' (%d MB)? [y/N] "+reset, target.Project, target.SizeMB))
		if strings.ToLower(confirm) == "y" {
			if err := os.RemoveAll(target.Path); err != nil {
				fmt.Println(red + "  Error: " + err.Error() + reset)
			} else {
				fmt.Printf(green+"  Deleted: %s\n"+reset, target.Path)
				results = append(results[:idx-1], results[idx:]...)
				if len(results) == 0 {
					fmt.Println(green + "\n  All node_modules cleared." + reset)
					prompt(dim + "  Press Enter to go back..." + reset)
					return
				}
			}
		}
	}
}