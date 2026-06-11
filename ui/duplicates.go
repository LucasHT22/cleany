package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cleany/scanner"
)

func viewDuplicates(root string) {
	minStr := prompt(bold + "  Min file size to check in MB (e.g. 10): " + reset)
	var minMB int64 = 10
	fmt.Sscanf(minStr, "%d", &minMB)
	if minMB < 1 {
		minMB = 1
	}

	fmt.Println()
	fmt.Println(dim + "  Scanning for duplicates (this may take a while)..." + reset)

	groups := scanner.FindDuplicates(root, minMB)

	clearScreen()
	PrintBanner()
	fmt.Println(bold + red + "  DUPLICATE FILES" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	if len(groups) == 0 {
		fmt.Println(green + "  No duplicates found." + reset)
		fmt.Println()
		prompt(dim + "  Press Enter to go back..." + reset)
		return
	}

	var totalWasted int64
	for _, g := range groups {
		totalWasted += g.Wasted
	}

	limit := 10
	if len(groups) < limit {
		limit = len(groups)
	}

	for i, g := range groups[:limit] {
		fmt.Printf("  %s%d.%s %s copies -  %s%s wasted%s %s(%s each)%s\n", bold+barColor(i), i+1, reset, fmt.Sprintf("%d", len(g.Files)), bold+red, fmtSize(g.Wasted), reset, dim, fmtSize(g.Size), reset)
		for _, f := range g.Files {
			rel := f
			if r, err := filepath.Rel(root, f); err == nil {
				rel = r
			}
			fmt.Printf("	%s%s%s\n", dim, truncate(rel, 65), reset)
		}
		fmt.Println()
	}

	fmt.Printf("  %sTotal wasted: %s%s%s across %d duplicate groups%s\n\n", dim, bold+white, fmtSize(totalWasted), reset, len(groups), reset)

	fmt.Println(dim + "  Enter a group number to delete duplicates (keeps first copy), or 'all', or 'q'." + reset)
	fmt.Println()

	for {
		choice := prompt(bold + "  > " + reset)
		if choice == "q" || choice == "Q" {
			return
		}

		if strings.ToLower(choice) == "all" {
			confirm := prompt(fmt.Sprintf(red+"  Delete all duplicates (~%s freed)? [y/N] "+reset, fmtSize(totalWasted)))
			if strings.ToLower(confirm) == "y" {
				deleted, failed := 0, 0
				for _, g := range groups[:limit] {
					for _, f := range g.Files[1:] {
						if err := os.Remove(f); err != nil {
							failed++
						} else {
							deleted++
						}
					}
				}
				fmt.Printf(green+"  Done! Deleted %d duplicate files."+reset+"\n", deleted)
				if failed > 0 {
					fmt.Printf(yellow+"  %d failed.\n"+reset, failed)
				}
				prompt(dim + "  Press Enter to go back..." + reset)
				return
			}
			continue
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx > limit {
			fmt.Println(red + "  Invalid choice." + reset)
			continue
		}

		g := groups[idx-1]
		fmt.Printf("  Keeping %s%s%s\n", green, g.Files[0], reset)
		fmt.Printf("  Deleting %d copies (%s):\n", len(g.Files)-1, fmtSize(g.Wasted))
		for _, f := range g.Files[1:] {
			fmt.Printf("	%s%s%s\n", dim, f, reset)
		}
		confirm := prompt(fmt.Sprintf(red+"  Confirm delete? [y/N] "+reset))
		if strings.ToLower(confirm) == "y" {
			for _, f := range g.Files[1:] {
				if err := os.Remove(f); err != nil {
					fmt.Printf(red+"  Failed: %s\n"+reset, f)
				} else {
					fmt.Printf(green+"  Deleted: %s\n"+reset, f)
				}
			}
			groups = append(groups[:idx-1], groups[idx:]...)
			if len(groups) == 0 {
				fmt.Println(green + "\n  All duplicates cleared!" + reset)
				prompt(dim + "  Press Enter to go back..." + reset)
				return
			}
		}
	}
}