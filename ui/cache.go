package ui

import (
	"fmt"
	"os"
	"strings"

	"cleany/scanner"
)

func viewCache() {
	fmt.Println(dim + "  Scanning known app caches..." + reset)

	results := scanner.ScanCaches()

	clearScreen()
	PrintBanner()
	fmt.Println(bold + cyan + "  APP CACHE FILES" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	if len(results) == 0 {
		fmt.Println(green + "  No app caches found! All clean:))" + reset)
		fmt.Println()
		prompt(dim + "  Press Enter to go back..." + reset)
		return
	}

	max := results[0].SizeMB
	for _, r := range results {
		if r.SizeMB > max {
			max = r.SizeMB
		}
	}
	if max == 0 {
		max = 1
	}

	for i, r := range results {
		filled := int(30 * float64(r.SizeMB) / float64(max))
		if filled < 1 {
			filled = 1
		}
		bar := barColor(i) + strings.Repeat("|", filled) + dim + strings.Repeat(".", 30-filled) + reset
		fmt.Printf("  %2d. %-20s %s  %s%d MB%s\n		%s%s%s\n\n", i+1, truncate(r.App, 20), bar, bold+barColor(i), r.SizeMB, reset, dim, truncate(r.Path, 60), reset)
	}

	var totalMB int64
	for _, r := range results {
		totalMB += r.SizeMB
	}

	fmt.Printf("  %sTotal recoverable: %s%d MB%s\n\n", dim, bold+white, totalMB, reset)
	fmt.Println(yellow + bold + "  ! Close any listed apps before deleting their cache!" + reset)
	fmt.Println(yellow + "		Deleting while an app is running can crash it or corrupt your project." + reset)
	fmt.Println()
	fmt.Println(dim + "  Enter a number to delete that cache, 'all' to delete everything;) or 'q' to go back." + reset)
	fmt.Println()

	for {
		choice := prompt(bold + "  > " + reset)

		if choice == "q" || choice == "Q" {
			return
		}

		if strings.ToLower(choice) == "all" {
			confirm := prompt(fmt.Sprintf(red+"  Delete ALL %d caches (~%d MB)? [y/N] "+reset, len(results), totalMB))
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
				fmt.Printf(green+"\n  Done. Deleted %d/%d cache folders.\n"+reset, deleted, len(results))
				prompt(dim + "  Press Enter to go back..." + reset)
				return
			}
			continue
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(results) {
			fmt.Println(red + "  Invalid choice." + reset)
			continue
		}

		target := results[idx-1]
		warning := ""
		if strings.Contains(target.App, "DaVinci") && strings.Contains(target.Path, "Optimized") {
			warning = yellow + "\n		Note: if your original footage is on an offline drive, this is your only playable copy.\n  " + reset
		}
		confirm := prompt(fmt.Sprintf(red+"  Delete %s cache (%d MB)?%s [y/N] " + reset, target.App, target.SizeMB, warning))
		if strings.ToLower(confirm) == "y" {
			if err := os.RemoveAll(target.Path); err != nil {
				fmt.Println(red + "  Error: " + err.Error() + reset)
			} else {
				fmt.Printf(green + "  Deleted: %s\n" + reset, target.Path)
				results = append(results[:idx-1], results[idx:]...)
				if len(results) == 0 {
					fmt.Println(green + "\n  All caches cleared." + reset)
					prompt(dim + "  Press Enter to go back..." + reset)
					return
				}
			}
		}
	}
}