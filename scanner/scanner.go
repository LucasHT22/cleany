package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	var()
	var totalSize int64
	var mu sync.Mutex

	rootEntry := &Entry{
		Path: path,
		Name: filepath.Base(root),
		IsDir: true,
	}

	dirSizes := make(map[string]int64)
	var dirSizesMu sync.Mutex

	var wg sync.WaitGroup
	fileChan := make(chan *Entry, 1000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.WalkDir(root, func(path string, do os.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
				return nil
			}
			info, err := d.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting info for %s: %v\n", path, err)
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
				fileChan <- e
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", root, err)
		}
		close(fileChan)
	}()

	var processorWg sync.WaitGroup
	processorWg.Add(1)
	go func() {
		defer processorWg.Done()
		for f := range fileChan {
			mu.Lock()
			totalSize += f.Size
			mu.Unlock()

			dirSizesMu.Lock()
			currentDir := filepath.Dir(f.Path)
			for strings.HasPrefix(currentDir, root) || currentDir == root {
				dirSizes[currentDir] += f.Size
				if currentDir == root {
					break
				}
				currentDir = filepath.Dir(currentDir)
			}
			dirSizesMu.Unlock()
		}
	}()

	wg.Wait()
	processorWg.Wait()

	var allEntries []*Entry
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
			Size: info.Size(),
		}
		if e.IsDir {
			dirSizesMu.Lock()
			e.Size = dirSizes[path]
			dirSizesMu.Unlock()
		}
		allEntries = append(allEntries, e)
		return nil
	})

	entryMap := make(map[string]*Entry)
	for _, e := range allEntries {
		entryMap[e.Path] = e
	}

	for _, e := range allEntries {
		if e.Path == root {
			rootEntry = ea

			continue
		}
		parentPath := filepath.Dir(e.Path)
		if parentEntry, ok := entryMap[parentPath]; ok {
			parentEntry.Children = append(parentEntry.Children, e)
			e.Parent = parentEntry
		}
	}

	for _, e := range allEntries {
		sort.Slice(e.Children, func(i, j int) bool {
			if e.Children[i].IsDir && !e.Children[j].IsDir {
				return true
			}
			if !e.Children[i].IsDir && e.Children[j].IsDir {
				return false
			}
			return e.Children[i].Size > e.Children[j].Size
		})
	}

	return rootEntry.Children, totalSize
}

func ExtBreakdown(root string) map[string]int64 {
	result := map[string]int64{}
	var mu sync.Mutex

	var mg sync.WaitGroup
	fileChan := make(chan *Entry, 1000)

	wg.Add(1)

	go func() {
		defer wg.Done()
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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
			fileChan <- &Entry{Ext: ext, Size: info.Size()}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory for extension breakdown %s: %v\n", root, err)
		}
		close(fileChan)
	}()

	var processorWg sync.WaitGroup
	processorWg.Add(1)
	go func() {
		defer processorWg.Done()
		for f := range fileChan {
			mu.Lock()
			result[f.Ext] += f.Size
			mu.Unlock()
		}
	}()

	wg.Wait()
	processorWg.Wait()

	return result
}

func LargestFiles(root string, n int) []*Entry {
	var all []*Entry

	var mu sync.Mutex
	var wg sync.WaitGroup
	fileChan := make(chan *Entry, 1000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			fileChan <- &Entry{
				Path: path,
				Name: d.Name(),
				Size: info.Size(),
				Ext: strings.ToLower(filepath.Ext(path)),
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory for largest files %s: %v\n", root, err)
		}
		close(fileChan)
	}()

	var processorWg sync.WaitGroup
	processorWg.Add(1)
	go func() {
		defer processorWg.Done()
		for f := range fileChan {
			mu.Lock()
			allFiles = append(allFiles, f)
			mu.Unlock()
		}
	}()

	wg.Wait()
	processorWg.Wait()

	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Size > allFiles[j].Size
	})
	if len(allFiles) > n {
		return allFiles[:n]
	}
	return allFiles
}