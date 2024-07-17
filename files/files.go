package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

type FindFilesFunc func(string, chan<- error) chan string

func FindRecursive(path string, errs chan<- error) chan string {
	var out = make(chan string)

	go func() {
		err := godirwalk.Walk(path, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if d, err := de.IsDirOrSymlinkToDir(); d == false && err != nil {
					out <- osPathname
				}
				return nil
			},
			ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
				errs <- fmt.Errorf("Error walking %s: %s", osPathname, err)
				return godirwalk.SkipNode
			},
			Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
		})
		if err != nil {
			errs <- err
		}
	}()

	return out
}

func FindFilesInRootDirectory(path string, errs chan<- error) chan string {
	var out = make(chan string)

	go func() {
		defer close(out)
		if entries, err := os.ReadDir(path); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && entry.Name()[0] != '.' {
					out <- filepath.Join(path, entry.Name())
				}
			}
		} else {
			errs <- err
		}
	}()

	return out
}
