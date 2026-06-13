package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cleany/scanner"
)

func viewExport(entries []*scanner.Entry, total int64, root string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	outPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "cleany_report_"+timestamp+".txt")

	fmt.Println(dim + "  Generating report..." + reset)

	var sb strings.Builder

	sb.WriteString("CLEANY DISK REPORT\n")
	sb.WriteString("Generated: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
	sb.WriteString("Root: " + root + "\n")
	sb.WriteString(strings.Repeat("=", 72) + "\n\n")

	sb.WriteString("DISK USAGE BY FOLDER\n")
	sb.WriteString(strings.Repeat("-", 72) + "\n")
	limit := 20
	if len(entries) < limit {
		limit = len(entries)
	}
	for i, e := range entries[:limit] {
		pct := 0.0
		if total > 0 {
			pct = float64(e.Size) / float64(total) * 100
		}
		sb.WriteString(fmt.Sprintf("  %2d. %-35s %10s  (%.1f%%)\n", i+1, e.Name, fmtSize(e.Size), pct))
	}
	sb.WriteString(fmt.Sprintf("\n  Total: %s\n\n", fmtSize(total)))

	sb.WriteString("TOP 20 LARGEST FILES\n")
	sb.WriteString(strings.Repeat("-", 72) + "\n")
	files := scanner.LargestFiles(root, 20)
	for i, f := range files {
		rel := f.Path
		if r, err := filepath.Rel(root, f.Path); err == nil {
			rel = r
		}
		sb.WriteString(fmt.Sprintf("  %2d. %-50s %10s\n", i+1, truncate(rel, 50), fmtSize(f.Size)))
	}
	sb.WriteString("\n")

	sb.WriteString("SPACE BY FILE TYPE\n")
	sb.WriteString(strings.Repeat("-", 72) + "\n")
	extMap := scanner.ExtBreakdown(root)
	type kv struct {
		k string
		v int64
	}
	var sorted []kv
	for k, v := range extMap {
		sorted = append(sorted, kv{k, v})
	}

	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if len(sorted) > 20 {
		sorted = sorted[:20]
	}
	for _, kv := range sorted {
		sb.WriteString(fmt.Sprintf("  %-15s %10s\n", kv.k, fmtSize(kv.v)))
	}
	sb.WriteString("\n")

	sb.WriteString("NODE_MODULES FOUND\n")
	sb.WriteString(strings.Repeat("-", 72) + "\n")
	nmResults := scanner.ScanNodeModules(root)
	if len(nmResults) == 0 {
		sb.WriteString("  None found.\n")
	} else {
		var nmTotal int64
		for _, r := range nmResults {
			sb.WriteString(fmt.Sprintf("  %-40s %10s\n", truncate(r.Path, 40), fmtSize(r.SizeMB*1024*1024)))
			nmTotal += r.SizeMB
		}
		sb.WriteString(fmt.Sprintf("\n  Total: %d MB\n", nmTotal))
	}
	sb.WriteString("\n")

	sb.WriteString("APP CACHES\n")
	sb.WriteString(strings.Repeat("-", 72) + "\n")
	caches := scanner.ScanCaches()
	if len(caches) == 0 {
		sb.WriteString("  None found.\n")
	} else {
		var cTotal int64
		for _, c := range caches {
			sb.WriteString(fmt.Sprintf("  %-25s %-40s %10s\n", c.App, truncate(c.Path, 40), fmtSize(c.SizeMB*1024*1024)))
			cTotal += c.SizeMB
		}
		sb.WriteString(fmt.Sprintf("\n  Total: %d MB\n", cTotal))
	}

	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		fmt.Println(red + "  Failed tp write report: " + err.Error() + reset)
	} else {
		fmt.Printf(green+"  Report saved to: %s\n"+reset, outPath)
	}

	fmt.Println()
	prompt(dim + "  Press Enter to go back..." + reset)
}