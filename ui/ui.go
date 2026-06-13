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

func viewFolders(entries []*scanner.Entry, total int64, root string) {
	currentEntries := entries
	currentTotal := total
	currentPath := root

	type level struct {
		path string
		entries []*scanner.Entry
		total int64
	}
	stack := []level{}

	for {
		clearScreen()
		PrintBanner()
		fmt.Println(bold + cyan + "  DISK USAGE BY FOLDER" + reset)
		fmt.Println(dim + "  " + strings.Repeat("-", 72) + reset)
		
		if len(stack) > 0 {
			crumb := ""
			for _, l := range stack {
				crumb += filepath.Base(l.path) + " > "
			}
			crumb += filepath.Base(currentPath)
			fmt.Printf("  %s%s%s\n", dim, crumb, reset)
		} else {
			fmt.Printf("  %s%s%s\n", dim, currentPath, reset)
		}
		fmt.Println()

		max := int64(1)
		if len(currentEntries) > 0 {
			max = currentEntries[0].Size
		}
		limit := 15
		if len(currentEntries) < limit {
			limit = len(currentEntries)
		}
		for i, e := range currentEntries[:limit] {
			pct := 0.0
			if currentTotal > 0 {
				pct = float64(e.Size) / float64(currentTotal) * 100
			}
			marker := " "
			if e.IsDir {
				marker = ">"
			}
			fmt.Printf("  %s%2d.%s%s %s\n", bold+barColor(i), i+1, reset, marker, barGraph(e.Name, e.Size, max, 28, barColor(i), pct)[2:])
		}
		fmt.Printf("\n %sTotal: %s%s%s\n", dim, bold+white, fmtSize(currentTotal), reset)
		
		if len(stack) > 0 {
			fmt.Println(dim + "  Enter a number to drill in, 'q' to go up, 'qq' to exit." + reset)
		} else {
			fmt.Println(dim + "  Enter a number to drill into a folder, or 'q' to go back." + reset)
		}
		fmt.Println()

		choice := prompt(bold + "  > " + reset)

		if choice == "qq" {
			return
		}
		if choice == "q" || choice == "Q" {
			if len(stack) == 0 {
				return
			}

			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			currentPath = top.path
			currentEntries = top.entries
			currentTotal = top.total
			continue
		}

		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > limit {
			continue
		}

		target := currentEntries[idx-1]
		if !target.IsDir {
			fmt.Println(yellow + "  That's a file, not a folder!" + reset)
			prompt("  Press Enter to continue...")
			continue
		}

		stack = append(stack, level{currentPath, currentEntries, currentTotal})
		fmt.Println(dim + "  Scanning..." + reset)
		subEntries, subTotal := scanner.Scan(target.Path)
		currentPath = target.Path
		currentEntries = subEntries
		currentTotal = subTotal
	}
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

	fmt.Println()
	ext := prompt(dim + "  Delete all files by extension? (e.g. .dvcc) or Enter to skip: " + reset)
	ext =strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	fmt.Printf(dim+"  Scanning for %s files...\n"+reset, ext)
	var matches []string
	var totalSize int64
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ext {
			info, err := d.Info()
			if err == nil {
				totalSize += info.Size()
			}
			matches = append(matches, path)
		}
		return nil
	})

	if len(matches) == 0 {
		fmt.Printf(yellow+"  No %s files found.\n"+reset, ext)
		return
	}

	fmt.Printf("\n  Found %s%d files%s - %s%s%s\n\n", bold+red, len(matches), reset, bold+white, fmtSize(totalSize), reset)

	confirm := prompt(fmt.Sprintf(red+ "  Delete all %d %s files (%s)? [y/N] "+reset, len(matches), ext, fmtSize(totalSize)))
	if strings.ToLower(confirm) !="y" {
		fmt.Println(dim + "  Cancelled." + reset)
		return
	}

	deleted, failed := 0, 0
	for _, p := range matches {
		if err := os.Remove(p); err != nil {
			failed++
		} else {
			deleted++
		}
	}

	fmt.Printf(green+"  Done. Deleted %d files."+reset, deleted)
	if failed > 0 {
		fmt.Printf(yellow+"  %d failed (permission delied?)"+reset, failed)
	}
	fmt.Println()
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
		fmt.Println(cyan + "  [7]" + reset + "  node_modules hunter")
		fmt.Println(magenta + "  [8]" + reset + "  Duplicate file finder")
		fmt.Println(yellow + "  [9]" + reset + "  Startup programs")
		fmt.Println(white + "  [0]" + reset + "  Export report to Desktop")
		fmt.Println(dim + "  [q]" + reset + "  Quit")
		fmt.Println()

		choice := prompt(bold + "  > " + reset)

		clearScreen()
		PrintBanner()

		switch choice {
		case "1":
			viewFolders(entries, total, root)
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
		case "7":
			viewNodeModules(root)
		case "8":
			viewDuplicates(root)
		case "9":
			viewExport(entries, total, root)
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