package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/anibaldeboni/rx/files"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(unzipCmd)
}

var unzipCmd = &cobra.Command{
	Use:     "unzip [path]",
	Short:   "inflate rom files.",
	Long:    `Inflate content files from zip files`,
	PreRunE: validateCmdPath,
	Run:     executeUnzip,
}

func executeUnzip(cmd *cobra.Command, args []string) {
	errs := make(chan error)
	defer close(errs)

	HandleErrors(errs)
	SetupWorkers(unzipWorker, options.FindFunc, args[0], errs)
}

type WorkerFunc func(*sync.WaitGroup, <-chan string, chan<- error)

func SetupWorkers(worker WorkerFunc, findFiles files.FindFilesFunc, path string, errs chan<- error) {
	var wg sync.WaitGroup
	files := findFiles(path, errs)
	wg.Add(options.Workers)
	for i := 0; i < options.Workers; i++ {
		go worker(&wg, files, errs)
	}

	wg.Wait()
}

func unzipWorker(wg *sync.WaitGroup, files <-chan string, errs chan<- error) {
	defer wg.Done()

	for file := range files {
		if filepath.Ext(file) != ZipExtension {
			continue
		}

		zipFile, err := zip.OpenReader(file)
		if err != nil {
			errs <- fmt.Errorf("Error opening zip archive %s: %w", file, err)
		}
		defer zipFile.Close()
		for _, zipContent := range zipFile.File {
			outputFilePath := filepath.Join(options.Output, zipContent.Name)

			log.Println("Extracting file", outputFilePath)

			if zipContent.FileInfo().IsDir() {
				if err := os.MkdirAll(outputFilePath, os.ModePerm); err != nil {
					errs <- fmt.Errorf("Error creating directory %s: %w", outputFilePath, err)
				}
				continue
			}

			if err := os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm); err != nil {
				errs <- fmt.Errorf("Error creating directory %s: %w", filepath.Dir(outputFilePath), err)
			}

			dstFile, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipContent.Mode())
			if err != nil {
				errs <- fmt.Errorf("Error creating file %s: %w", outputFilePath, err)
			}

			srcFile, err := zipContent.Open()
			if err != nil {
				errs <- fmt.Errorf("Error compressed file %s: %w", zipContent.Name, err)
			}
			if _, err := io.Copy(dstFile, srcFile); err != nil {
				errs <- fmt.Errorf("Error inflating file %s: %w", zipContent.Name, err)
			}

			dstFile.Close()
			srcFile.Close()
		}
	}
}
