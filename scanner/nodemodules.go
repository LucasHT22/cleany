package scanner

import (
	"os"
	"path/filepath"
	"sort"
)

type NodeModulesResult struct {
	Path string
	Project string
	SizeMB int64
}

func ScanNodeModules(root string) []NodeModulesResult {
	var results []NodeModulesResult

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == "node_modules" {
			size := scanDirSize(path)
			project := filepath.Base(filepath.Dir(path))
			results = append(results, NodeModulesResult{
				Path: path,
				Project: project,
				SizeMB: size / 1024 / 1024,
			})
			return filepath.SkipDir
		}
		return nil
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].SizeMB > results[j].SizeMB
	})

	return results
}