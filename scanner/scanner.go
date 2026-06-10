package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Path string
	Name string
	Size int64
	IsDir bool
	Ext string
	Children []*Entry
	Parent *Entry
}

func Scan(root string) ([]*Entry, int64) {
	dirSizes := map[string]int64{}
	fileMap := map[string]*Entry{}
	var allFiles []*Entry
	var total int64

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		e := &Entry{
			Path: path,
			Name: d.Name(),
			IsDir: d.IsDir(),
			Ext: ext,
		}

		if !d.IsDir() {
			e.Size = info.Size()
			total += e.Size
			allFiles = append(allFiles, e)

			dir := filepath.Dir(path)
			for dir != "" && dir != "." {
				dirSizes[dir] += e.Size
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}

		fileMap[path] = e
		return nil
	})

	rootEntry := &Entry{
		Path: root,
		Name: root,
		IsDir: true,
		Size: dirSizes[root],
	}

	seen := map[string]bool{}
	for _, f := range allFiles {
		rel, _ := filepath.Rel(root, f.Path)
		parts := strings.SplitN(rel, string(os.PathSeparator), 2)
		if len(parts) == 0 {
			continue
		}
		topName := parts[0]
		topPath := filepath.Join(root, topName)
		if !seen[topPath] {
			seen[topPath] = true
			info, err := os.Stat(topPath)
			if err != nil {
				continue
			}
			child := &Entry{
				Path: topPath,
				Name: topName,
				IsDir: info.IsDir(),
				Parent: rootEntry,
			}
			if !info.IsDir() {
				child.Size = info.Size()
			}
			rootEntry.Children = append(rootEntry.Children, child)
		}
	}

	sort.Slice(rootEntry.Children, func(i, j int) bool {
		return rootEntry.Children[i].Size > rootEntry.Children[j].Size
	})
	if len(allFiles) > 20 {
		allFiles = allFiles[:20]
	}

	extMap := map[string]int64{}
	for _, f := range allFiles {
		if f.Ext == "" {
			extMap["(no ext)"] += f.Size
		} else {
			extMap[f.Ext] += f.Size
		}
	}

	_ = extMap

	return rootEntry.Children, total
}

func ExtBreakdown(root string) map[string]int64 {
	result := map[string]int64{}
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = "(no ext)"
		}
		result[ext] += info.Size()
		return nil
	})
	return result
}

func LargestFiles(root string, n int) []*Entry {
	var all []*Entry
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		all = append(all, &Entry{
			Path: path,
			Name: d.Name(),
			Size: info.Size(),
			Ext: strings.ToLower(filepath.Ext(path)),
		})
		return nil
	})
	sort.Slice(all, func(i, j int) bool {
		return all[i].Size > all[j].Size
	})
	if len(all) > n {
		return all[:n]
	}
	return all
}