package scanner

import (
	"os"
	"path/filepath"
)

type CacheTarget struct {
	App string
	Label string
	Paths []string
}

type CacheResult struct {
	App string
	Label string
	Path string
	SizeMB int64
	Exists bool
}

func knownCaches() []CacheTarget {
	appdata := os.Getenv("APPDATA")
	localappdata := os.Getenv("LOCALAPPDATA")
	programdata := os.Getenv("PROGRAMDATA")
	userprofile := os.Getenv("USERPROFILE")

	return []CacheTarget{
		{
			App: "DaVinci Resolve",
			Label: "Optimized media & cache",
			Paths: []string{
				filepath.Join(userprofile, "AppData", "Roaming", "Blackmagic Design", "DaVinci Resolve", "Support", "CacheClip"),
				filepath.Join(userprofile, "AppData", "Roaming", "Blackmagic Design", "DaVinci Resolve", "Support", "Optimized Media"),
				filepath.Join(programdata, "Blackmagic Design", "DaVinci Resolve", "Support", "CacheClip"),
				filepath.Join(programdata, "Blackmagic Design", "DaVinci Resolve", "Support", "Optimized Media"),
			},
		},
		{
			App: "Adobe Premiere",
			Label: "Media cache & previews",
			Paths: []string{
				filepath.Join(localappdata, "Adobe", "Premiere Pro", "Media Cache"),
				filepath.Join(localappdata, "Adobe", "Premiere Pro", "Media Cache Files"),
				filepath.Join(localappdata, "Adobe", "Premiere Pro", "Preview Files"),
				filepath.Join(appdata, "Adobe", "Common", "Media Cache"),
				filepath.Join(appdata, "Adobe", "Common", "Media Cache Files"),
			},
		},
		{
			App: "Adobe After Effects",
			Label: "Disk cache",
			Paths: []string{
				filepath.Join(localappdata, "Adobe", "After Effects", "Disk Cache"),
				filepath.Join(localappdata, "Adobe", "After Effects"),
			},
		},
		{
			App: "Adobe Photoshop",
			Label: "Scratch & temp files",
			Paths: []string{
				filepath.Join(localappdata, "Adobe", "Photoshop", "CT Font Cache"),
				filepath.Join(localappdata, "Temp"),
			},
		},
		{
			App: "Adobe Lightroom",
			Label: "Preview cache",
			Paths: []string{
				filepath.Join(appdata, "Adobe", "Lightroom", "Caches"),
				filepath.Join(localappdata, "Adobe", "Lightroom", "Caches"),
			},
		},
		{
			App: "Adobe Creative Cloud",
			Label: "CC cache",
			Paths: []string{
				filepath.Join(localappdata, "Adobe", "CoreSync", "cache"),
				filepath.Join(localappdata, "Adobe", "Creative Cloud Libraries", "cache"),
			},
		},
		{
			App: "Spotify",
			Label: "Audio cache",
			Paths: []string{
				filepath.Join(localappdata, "Spotify", "Data"),
				filepath.Join(localappdata, "Spotify", "Storage"),
			},
		},
		{
			App: "Discord",
			Label: "App & image cache",
			Paths: []string{
				filepath.Join(appdata, "discord", "Cache"),
				filepath.Join(appdata, "discord", "Code Cache"),
				filepath.Join(appdata, "discord", "GPUCache"),
			},
		},
		{
			App: "Slack",
			Label: "App cache",
			Paths: []string{
				filepath.Join(appdata, "Slack", "Cache"),
				filepath.Join(appdata, "Slack", "Code Cache"),
			},
		},
		{
			App: "Microsoft Teams",
			Label: "App & media cache",
			Paths: []string{
				filepath.Join(appdata, "Microsoft", "Teams", "Cache"),
				filepath.Join(appdata, "Microsoft", "Teams", "blob_storage"),
				filepath.Join(appdata, "Microsoft", "Teams", "GPUCache"),
				filepath.Join(appdata, "Microsoft", "Teams", "databases"),
				filepath.Join(appdata, "Microsoft", "Teams", "Local Storage"),
			},
		},
		{
			App: "Google Chrome",
			Label: "Browser cache",
			Paths: []string{
				filepath.Join(localappdata, "Google", "Chrome", "User Data", "Default", "Cache"),
				filepath.Join(localappdata, "Google", "Chrome", "User Data", "Default", "Code Cache"),
				filepath.Join(localappdata, "Google", "Chrome", "User Data", "Default", "GPUCache"),
			},
		},
		{
			App: "Firefox",
			Label: "Browser cache",
			Paths: []string{
				filepath.Join(localappdata, "Mozilla", "Firefox", "Profiles"),
			},
		},
		{
			App: "Steam",
			Label: "Download cache & depots",
			Paths: []string{
				`C:\Program Files (x86)\Steam\steamapps\downloading`,
				`C:\Program Files (x86)\Steam\steamapps\temp`,
				`C:\Program Files\Steam\steamapps\downloading`,
				`C:\Program Files\Steam\steamapps\temp`,
			},
		},
		{
			App: "Epic Games",
			Label: "Download cache",
			Paths: []string{
				filepath.Join(localappdata, "EpicGamesLauncher", "Saved", "webcache"),
				filepath.Join(localappdata, "EpicGamesLauncher", "Saved", "Logs"),
			},
		},
	}
}

func ScanCaches() []CacheResult {
	targets := knownCaches()
	var results []CacheResult

	for _, t := range targets {
		for _, p := range t.Paths {
			if _, err := os.Stat(p); os.IsNotExist(err) {
				continue
			}
			size := scanDirSize(p)
			if size == 0 {
				continue
			}
			results = append(results, CacheResult{
				App: t.App,
				Label: t.Label,
				Path: p,
				SizeMB: size / 1024 / 1024,
				Exists: true,
			})
		}
	}

	return results
}