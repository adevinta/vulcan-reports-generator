/*
Copyright 2021 Adevinta
*/

package files

import (
	"os"
	"path/filepath"
)

// ListDirFiles returns the file path for every
// regular file contained in the specified directory.
// Search is performed recursively with no depth limit.
func ListDirFiles(dirPath string) ([]string, error) {
	var files []string

	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	fii, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, f := range fii {
		if f.Mode().IsRegular() {
			filePath := filepath.Join(dirPath, f.Name())
			files = append(files, filePath)
		} else if f.IsDir() {
			subDirPath := filepath.Join(dirPath, f.Name())
			subDirFiles, err := ListDirFiles(subDirPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subDirFiles...)
		}
	}

	return files, nil
}
