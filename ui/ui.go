package ui

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cleany/scanner"
)

const (
	reset = "\033[0m"
	bold = "\033[1m"
	dim = "\033[2m"
	red = "\033[31m"
	green = "\033[32m"
	yellow = "\033[33m"
	blue = "\033[34m"
	magenta = "\033[35m"
	cyan = "\033[36m"
	white = "\033[37m"
	bgRed = "\033[41m"
)

var barColors = []string{red, magenta, yellow, cyan, green, blue, white}

func barColor(i int) string {
	return barColors[i%len(barColors)]
}

func PrintBanner() {
	fmt.Println(cyan + bold)
	fmt.Println("  ██████╗██╗     ███████╗ █████╗ ███╗   ██╗██╗   ██╗")
	fmt.Println(" ██╔════╝██║     ██╔════╝██╔══██╗████╗  ██║╚██╗ ██╔╝")
	fmt.Println(" ██║     ██║     █████╗  ███████║██╔██╗ ██║ ╚████╔╝ ")
	fmt.Println(" ██║     ██║     ██╔══╝  ██╔══██║██║╚██╗██║  ╚██╔╝  ")
	fmt.Println(" ╚██████╗███████╗███████╗██║  ██║██║ ╚████║   ██║   ")
	fmt.Println("  ╚═════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝  ")
	fmt.Println(reset)
	fmt.Println(dim + " disk space visualizer & cleaner" + reset)
	fmt.Println()
}

func fmtSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func barGraph(label string, val, max int64, width int, color string, pct float64) string {
	filled := int(math.Round(float64(width) * float64(val) / float64(max)))
	if filled < 1 && val > 0 {
		filled = 1
	}
	bar := color + strings.Repeat("█", filled) + dim + strings.Repeat("░", width-filled) + reset
	return fmt.Sprintf(" %-30s %s %s %s(%.1f%%)%s", truncate(label, 30), bar, color+bold+fmtSize(val)+reset, dim, pct, reset)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "~"
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func prompt(msg string) string {
	fmt.Print(msg)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return strings.TrimSpace(sc.Text())
}

func viewFolders(entries []*scanner.Entry, total int64) {
	fmt.Println(bold + cyan + "  DISK USAGE BY FOLDER" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	max := int64(1)
	if len(entries) > 0 {
		max = entries[0].Size
	}
	limit := 15
	if len(entries) < limit {
		limit = len(entries)
	}
	for i, e := range entries[:limit] {
		pct := 0.0
		if total > 0 {
			pct = float64(e.Size) / float64(total) * 100
		}
		fmt.Println(barGraph(e.Name, e.Size, max, 30, barColor(i), pct))
	}
	fmt.Printf("\n %sTotal: %s%s%s\n", dim, bold+white, fmtSize(total), reset)
}

func viewLargestFiles(root string) {
	fmt.Println(bold + magenta + "  TOP 15 LARGEST FILES" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	files := scanner.LargestFiles(root, 15)
	if len(files) == 0 {
		fmt.Println("  No files found.")
		return
	}
	max := files[0].Size
	for i, f := range files {
		pct := float64(f.Size) / float64(max) * 100
		rel := f.Path
		if r, err := filepath.Rel(root, f.Path); err == nil {
			rel = r
		}
		fmt.Println(barGraph(rel, f.Size, max, 30, barColor(i), pct))
	}
}

func viewExtBreakdown(root string) {
	fmt.Println(bold + yellow + "  SPACE BY FILE TYPE" + reset)
	fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
	extMap := scanner.ExtBreakdown(root)
	type kv struct {
		k string
		v int64
	}
	var sorted []kv
	var total int64
	for k, v := range extMap {
		sorted = append(sorted, kv{k, v})
		total += v
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	if len(sorted) > 15 {
		sorted = sorted[:15]
	}
	max := int64(1)
	if len(sorted) > 0 {
		max = sorted[0].v
	}
	for i, kv := range sorted {
		pct := 0.0
		if total > 0 {
			pct = float64(kv.v) / float64(total) * 100
		}
		fmt.Println(barGraph(kv.k, kv.v, max, 30, barColor(i), pct))
	}
}

func viewDelete(root string) {
	files := scanner.LargestFiles(root, 30)
	if len(files) == 0 {
		fmt.Println("  No files found.")
		return
	}

	for {
		clearScreen()
		PrintBanner()
		fmt.Println(bold + red + "  DELETE FILES" + reset)
		fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
		fmt.Println(dim + "  Enter a number to delete that file, or 'q' to go back." + reset)
		fmt.Println()

		max := files[0].Size
		for i, f := range files {
			pct := float64(f.Size) / float64(max) * 100
			rel := f.Path
			if r, err := filepath.Rel(root, f.Path); err == nil {
				rel = r
			}
			idx := fmt.Sprintf("%2d. ", i+1)
			fmt.Print("  " + dim + idx + reset)
			fmt.Println(barGraph(rel, f.Size, max, 25, barColor(i), pct))
		}

		fmt.Println()
		choice := prompt(bold + "  > " + reset)
		if choice == "q" || choice == "Q" {
			return
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(files) {
			fmt.Println(red + "  Invalid choice." + reset)
			prompt("  Press enter to continue...")
			continue
		}

		target := files[idx-1]
		confirm := prompt(fmt.Sprintf(red+"  Delete %s (%s)? [y/N] "+reset, target.Name, fmtSize(target.Size)))
		if strings.ToLower(confirm) == "y" {
			if err := os.Remove(target.Path); err != nil {
				fmt.Println(red + "  Error: " + err.Error() + reset)
			} else {
				fmt.Println(green + "  Deleted: " + target.Path + reset)
				files = append(files[:idx-1], files[idx:]...)
				if len(files) == 0 {
					prompt("  No more files. Press Enter...")
					return
				}
			}
			prompt("  Press Enter to continue...")
		}
	}
}

func Run(entries []*scanner.Entry, total int64, root string) {
	for {
		clearScreen()
		PrintBanner()

		fmt.Println(bold + "  Hello! What do you want to see?" + reset)
		fmt.Println()
		fmt.Println(cyan + "  [1]" + reset + "  Disk usage by folder")
		fmt.Println(magenta + "  [2]" + reset + "  Largest files")
		fmt.Println(yellow + "  [3]" + reset + "  Space by file type")
		fmt.Println(red + "  [4]" + reset + "  Delete files")
		fmt.Println(green + "  [5]" + reset + "  Unused apps + combo suggestions")
		fmt.Println(blue + "  [6]" + reset + "  Clean app caches")
		fmt.Println(dim + "  [q]" + reset + "  Quit")
		fmt.Println()

		choice := prompt(bold + "  > " + reset)

		clearScreen()
		PrintBanner()

		switch choice {
		case "1":
			viewFolders(entries, total)
			fmt.Println()
			prompt(dim + "  Press enter to go back..." + reset)
		case "2":
			viewLargestFiles(root)
			fmt.Println()
			prompt(dim + "  Press enter to go back..." + reset)
		case "3":
			viewExtBreakdown(root)
			fmt.Println()
			prompt(dim + "  Press enter to go back..." + reset)
		case "4":
			viewDelete(root)
		case "5":
			viewApps(root)
		case "6":
			viewCache()
		case "q", "Q":
			clearScreen()
			fmt.Println(cyan + bold + "\n Bye bye!\n" + reset)
			return
		default:
			fmt.Println(red + "  Unknown option." + reset)
			prompt("  Press Enter...")
		}
	}
}