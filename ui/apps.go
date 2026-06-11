package ui

import (
	"fmt"
	"strings"

	"cleany/scanner"
)

func viewApps(root string) {
	days := 0
	for days <= 0 {
		input := prompt(bold + "  Unused threshold (days, numbers only): " + reset)
		fmt.Sscanf(input, "%d", &days)
		if days <= 0 {
			fmt.Println(red + "  Enter a number greater than 0." + reset)
		}
	}

	fmt.Println()
	fmt.Println(dim + "  Scanning installed apps..." + reset)

	apps, err := scanner.ScanApps(days)
	if err != nil {
		fmt.Println(red + "  Failed to read registry: " + err.Error() + reset)
		prompt("  Press Enter to go back...")
		return
	}
	if len(apps) == 0 {
		fmt.Printf(green + "  No apps unused for %d+ days found.\n"+reset, days)
		prompt("  Press Enter to go back...")
		return
	}

	clearScreen()
	PrintBanner()
	fmt.Printf(bold+cyan+"  UNUSED APPS  (not used in %d+ days)\n"+reset, days)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	max := apps[0].SizeMB
	if max == 0 {
		max = 1
	}
	for i, a := range apps {
		pct := 0.0
		if max > 0 {
			pct = float64(a.SizeMB) / float64(max) * 100
		}
		sizeLabel := fmt.Sprintf("%d MB", a.SizeMB)
		if a.SizeMB == 0 {
			sizeLabel = "? MB"
		}
		filled := int(30 * float64(a.SizeMB) / float64(max))
		if filled < 1 && a.SizeMB > 0 {
			filled = 1
		}
		bar := barColor(i) + strings.Repeat("|", filled) + dim + strings.Repeat(".", 30-filled) + reset
		daysStr := fmt.Sprintf("%dd ago", a.DaysSinceUse)
		if a.DaysSinceUse == 0 {
			daysStr = "unknown"
		}
		fmt.Printf("  %-28s %s %s%-8s%s  %s[%s / %s]%s\n", truncate(a.Name, 28), bar, barColor(i)+bold, sizeLabel, reset, dim, a.Category, daysStr, reset,)
		_ = pct
	}

	var TotalMB int64
	for _, a := range apps {
		TotalMB += a.SizeMB
	}
	fmt.Printf("\n %s%d apps ~%d MB recoverable%s\n", dim, len(apps), TotalMB, reset)
	fmt.Println()
	prompt(dim + "  Press Enter to see combo suggestions..." + reset)

	clearScreen()
	PrintBanner()
	combos := scanner.SuggestCombos(apps)

	fmt.Println(bold + yellow + "  COMBO SUGGESTIONS (multiple unused apps in same category)" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	if len(combos) == 0 {
		fmt.Println(dim + "  No combo suggestions - unused apps are all in different categories." + reset)
	} else {
		for i, g := range combos {
			fmt.Printf("  %s%s%s  %s(~%d MB total)%s\n", bold+barColor(i), g.Category, reset, dim, g.TotalMB, reset,)
			for _, a := range g.Apps {
				fmt.Printf("   %s%-28s%s %s%d MB  •  %dd unused%s\n", white, truncate(a.Name, 28), reset, dim, a.SizeMB, a.DaysSinceUse, reset,)
			}
			fmt.Println()
		}
	}
	fmt.Println(bold + magenta + "  BIGGEST SINGLE TARGETS" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	limit := 5
	if len(apps) < limit {
		limit = len(apps)
	}
	for _, a := range apps[:limit] {
		fmt.Printf("  %-30s  %s%d MB%s %s%s  •  %dd unused%s\n", truncate(a.Name, 30), bold+red, a.SizeMB, reset, dim, a.Category, a.DaysSinceUse, reset,)
	}

	fmt.Println()
	prompt(dim + "  Press Enter to go back..." + reset)
}

