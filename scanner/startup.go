package scanner

import (
	"os/exec"
	"strings"
)

type StartupEntry struct {
	Name string
	Command string
	Source string
}

func ScanStartup() ([]StartupEntry, error) {
	script := `
$entries = @()
 
$keys = @(
    @{Path='HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run'; Source='HKLM Run'},
    @{Path='HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run'; Source='HKCU Run'},
    @{Path='HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce'; Source='HKLM RunOnce'},
    @{Path='HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce'; Source='HKCU RunOnce'}
)
 
foreach ($key in $keys) {
    $props = Get-ItemProperty -Path $key.Path -ErrorAction SilentlyContinue
    if ($props) {
        $props.PSObject.Properties | Where-Object { $_.Name -notlike 'PS*' } | ForEach-Object {
            Write-Output ($_.Name + '|' + $_.Value + '|' + $key.Source)
        }
    }
}
	`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var entries []StartupEntry
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}
		entries = append(entries, StartupEntry{
			Name: strings.TrimSpace(parts[0]),
			Command: strings.TrimSpace(parts[1]),
			Source: strings.TrimSpace(parts[2]),
		})
	}
	return entries, nil
}

func DisableStartupEntry(name, source string) error {
	var keyPath string
	switch source {
	case "HKLM Run":
		keyPath = `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	case "HKCU Run":
		keyPath = `HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	case "HKLM RunOnce":
		keyPath = `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`
	case "HKCU RunOnce":
		keyPath = `HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`
	default:
		keyPath = `HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	}

	script := `Remove-ItemProperty -Path '` + keyPath + `' -Name '` + name + `' -ErrorAction Stop`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run()
}