package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"runtime"
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
	var ()
	fileChan = make(chan *Entry)
	var total int64
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
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
				fileChan <- 0
			}
			return nil
		})
		close(fileChan)
	}()

	wg.Wait()

	dirSizes := make(map[string]int64)
	for _, f := range allFiles {
		dir := filepath.Dir(f.Path)
		for dir != "" && dir != "." && strings.HasPrefix(dir, root) {
			dirSizes[dir] += f.Size
			dir = filepath.Dir(dir)
		}
		dirSizes[root] += f.Size
	}

	rootEntry := &Entry{
		Path: root,
		Name: filepath.Base(root),
		IsDir: true,
		Size: dirSizes[root],
	}

	seen := make(map[string]bool)
	for _, f := range allFiles {
		rel, _ := filepath.Rel(root, f.Path)
		if err != nil {
			continue
		}
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
		if rootEntry.Children[i].IsDir && !rootEntry.Children[j].IsDir {
			return true
		}
		if !rootEntry.Children[i].IsDir && rootEntry.Children[j].IsDir {
			return false
		}
		return rootEntry.Children[i].Size > rootEntry.Children[j].Size
	})
	
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