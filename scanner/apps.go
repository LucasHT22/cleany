package scanner

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type App struct {
	Name string
	Publisher string
	InstallDir string
	SizeMB int64
	InstallDate time.Time
	LastUsed time.Time
	DaysSinceUse int
	Category string
	UninstallCmd string
}

var categories = map[string][]string{
	"Game Launcher": {"steam", "epic", "gog", "battle.net", "ea app", "ubisoft", "origin", "itch"},
	"Adobe": {"adobe", "photoshop", "illustrator", "premiere", "acrobat", "lightroom", "after effects"},
	"Microsoft": {"microsoft", "office", "word", "excel", "powerpoint", "teams", "onedrive", "visual studio"},
	"Browser": {"chrome", "firefox", "opera", "brave", "edge", "vivaldi"},
	"Media": {"vlc", "spotify", "itunes", "winamp", "foobar", "kodi", "plex", "obs"},
	"Dev Tool": {"git", "node", "python", "java", "docker", "vscode", "jetbrains", "android studio", "postman"},
	"Utility": {"ccleaner", "7-zip", "winrar", "everything", "autoruns", "cpu-z", "hwinfo"},
}

func categorize(name string) string {
	lower := strings.ToLower(name)
	for cat, keywords := range categories {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return cat
			}
		}
	}
	return "Other"
}

func scanDirSize(dir string) int64 {
	var total int64
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.IsDir()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func lastUsedTime(dir string) time.Time {
	if dir == "" {
		return time.Time{}
	}
	var latest time.Time
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".exe" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return latest
}

func ScanApps(dayThreshold int) ([]*App, error) {
	script := `
$keys = @(
    'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
    'HKLM:\Software\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*',
    'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*'
)
$apps = @()
foreach ($key in $keys) {
    $apps += Get-ItemProperty $key -ErrorAction SilentlyContinue |
        Where-Object { $_.DisplayName -and $_.DisplayName -ne '' } |
        Select-Object DisplayName, Publisher, InstallLocation, EstimatedSize, InstallDate, UninstallString
}
$apps | ForEach-Object {
    $line = ($_.DisplayName -replace '|','') + '|' +
            ($_.Publisher -replace '|','') + '|' +
            ($_.InstallLocation -replace '|','') + '|' +
            ($_.EstimatedSize) + '|' +
            ($_.InstallDate) + '|' +
            ($_.UninstallString -replace '|','')
    Write-Output $line
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	appMap := map[string]*App{}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		if name == "" {
			continue
		}
		if _, exists := appMap[name]; exists {
			continue
		}

		publisher := strings.TrimSpace(parts[1])
		installDir := strings.TrimSpace(parts[2])
		sizeKB, _ := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
		installDateStr := strings.TrimSpace(parts[4])
		uninstallCmd := strings.TrimSpace(parts[5])

		var installDate time.Time
		if len(installDateStr) == 8 {
			installDate, _ = time.Parse("20060102", installDateStr)
		}

		sizeMB := sizeKB / 1024
		if sizeMB == 0 && installDir != "" {
			sizeMB = scanDirSize(installDir) / 1024 / 1024
		}

		lastUsed := lastUsedTime(installDir)
		daysOld := 0
		if !lastUsed.IsZero() {
			daysOld = int(math.Round(now.Sub(lastUsed).Hours() / 24))
		} else if !installDate.IsZero() {
			daysOld = int(math.Round(now.Sub(installDate).Hours() / 24))
		}

		app := &App{
			Name: name,
			Publisher: publisher,
			InstallDir: installDir,
			LastUsed: lastUsed,
			DaysSinceUse: daysOld,
			Category: categorize(name),
			UninstallCmd: uninstallCmd,
		}
		appMap[name] = app
	}

	var apps []*App
	for _, a := range appMap {
		if a.DaysSinceUse >= dayThreshold {
			apps = append(apps, a)
		}
	}

	sort.Slice(apps, func(i, j int) bool {
		return apps[i].SizeMB > apps[j].SizeMB
	})

	return apps, nil
}

type ComboGroup struct {
	Category string
	Apps []*App
	TotalMB int64
}

func SuggestCombos(apps []*App) []ComboGroup {
	catMap := map[string][]*App{}
	for _, a := range apps {
		catMap[a.Category] = append(catMap[a.Category], a)
	}

	var groups []ComboGroup
	for cat, list := range catMap {
		if len(list) < 2 {
			continue
		}
		var total int64
		for _, a := range list {
			total += a.sizeMB
		}
		groups = append(groups, ComboGroup{
			Category: cat,
			Apps: app,
			TotalMB: total,
		})
	}
	
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalMB > groups[j].TotalMB
	})

	return groups
}