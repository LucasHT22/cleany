package ui

import (
	"fmt"
	"strings"

	"cleany/scanner"
)

func viewStartup() {
	fmt.Println(dim + "  Reading startup programs..." + reset)

	entries, err := scanner.ScanStartup()
	if err != nil {
		fmt.Println(red + "  Failed to read startup entries: " + err.Error() + reset)
		prompt("  Press Enter to go back...")
		return
	}

	clearScreen()
	PrintBanner()
	fmt.Println(bold + yellow + "  STARTUP PROGRAMS" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	fmt.Println()

	if len(entries) == 0 {
		fmt.Println(dim + "  No startup entries found." + reset)
		prompt("  Press Enter to go back...")
		return
	}

	for i, e := range entries {
		fmt.Printf("  %s%2d.%s %-28s %s%s%s\n		%s%s%s\n\n", bold+barColor(i), i+1, reset, truncate(e.Name, 28), dim, e.Source, reset, dim, truncate(e.Command, 60), reset)
	}

	fmt.Println(yellow + "  ! Disabling removes the registry entry - the app stays installed." + reset)
	fmt.Println(yellow + "    You can re-enable it from the app's own settings." + reset)
	fmt.Println()
	fmt.Println(dim + "  Enter a number to disable that startup entry, or 'q' to go back." + reset)
	fmt.Println()

	for {
		choice := prompt(bold + "  > " + reset)
		if choice == "q" || choice == "Q" {
			return
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(entries) {
			fmt.Println(red + "  Invalid choice." + reset)
			continue
		}

		target := entries[idx-1]
		confirm := prompt(fmt.Sprintf(red+"  Disable '%s' from startup? [y/N] "+reset, target.Name))
		if strings.ToLower(confirm) == "y" {
			if err := scanner.DisableStartupEntry(target.Name, target.Source); err != nil {
				fmt.Println(red + "  Error: " + err.Error() + reset)
				fmt.Println(dim + "  Try running as Administrator!" + reset)
			} else {
				fmt.Printf(green+"  Disabled: %s\n"+reset, target.Name)
				entries = append(entries[:idx-1], entries[idx:]...)
				if len(entries) == 0 {
					fmt.Println(green + "\n  No more startup entries." + reset)
					prompt(dim + "  Press Enter to go back..." + reset)
					return
				}
			}
		}
	}
}