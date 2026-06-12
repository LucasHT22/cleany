package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type DuplicateGroup struct {
	Hash string
	Size int64
	Files []string
	Wasted int64
}

const partialHashBlockSize = 4096

func hashFilePartial(path string, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()

	buffer := make([]byte, partialHashBlockSize)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	h.Write(buffer[:n])

	if size > 2*partialHashBlockSize {
		_, err = f.Seek(-partialHashBlockSize, io.SeekEnd)
		if err != nil {
			return "", err
		}
		n, err = f.Read(buffer)
		if err != nil && err != io.EOF {
			return "", err
		}
		h.Write(buffer[:n])
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func FindDuplicates(root string, minSizeMB int64) []DuplicateGroup {
	minBytes := minSizeMB * 1024 * 1024
	sizeMap := map[int64][]string{}

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() < minBytes {
			return nil
		}
		sizeMap[info.Size()] = append(sizeMap[info.Size()], path)
		return nil
	})

	partialHashMap := map[string][]string{}
	for _, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			h, err := hashFilePartial(p, info.Size())
			if err != nil {
				continue
			}
			partialHashMap[h] =  append(partialHashMap[h], p)
		}
	}

	finalHashMap := map[string][]string{}
	for _, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}
		for _, p := range paths {
			h, err := hashFile(p)
			if err != nil {
				continue
			}
			finalHashMap[h] = append(finalHashMap[h], p)
		}
	}

	var groups []DuplicateGroup
	far hash, paths := range finalHashMap {
		if len(paths) < 2 {
			continue
		}
		info, err := os.Stat(paths[0])
		if err != nil {
			continue
		}
		size := info.Size()
		groups = append(groups, DuplicateGroup{
			Hash: hash,
			Size: size,
			Files: paths,
			Wasted: size * int64(len(paths)-1),
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Wasted > groups[j].Wasted
	})

	return groups
}