package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

const ZipExtension = ".zip"

func init() {
	rootCmd.AddCommand(zipCmd)
}

var zipCmd = &cobra.Command{
	Use:     "zip [path]",
	Short:   "compress rom files.",
	Long:    `Compress rom files into individual zip files.`,
	PreRunE: validateCmdPath,
	Run:     executeZip,
}

func executeZip(cmd *cobra.Command, args []string) {
	errs := make(chan error)
	defer close(errs)

	HandleErrors(errs)
	SetupWorkers(zipWorker, options.FindFunc, args[0], errs)

	fmt.Println("Done!")
}

func zipWorker(wg *sync.WaitGroup, files <-chan string, errs chan<- error) {
	defer wg.Done()

	for file := range files {
		func() {
			outputFileName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)) + ZipExtension
			outputFilePath := filepath.Join(options.Output, outputFileName)

			outputFile, err := os.Create(outputFilePath)
			defer outputFile.Close()
			if err != nil {
				errs <- fmt.Errorf("Error creating file %s: %w", outputFilePath, err)
				return
			}

			inputFile, err := os.Open(file)
			defer inputFile.Close()
			if err != nil {
				errs <- fmt.Errorf("Error opening file %s: %w", file, err)
				return
			}

			zipWriter := zip.NewWriter(outputFile)

			fileWriter, err := zipWriter.Create(filepath.Base(file))
			if err != nil {
				errs <- fmt.Errorf("Error creating file %s in zip: %w", file, err)
				return
			}

			fmt.Printf("Zipping %s\n", file)

			if _, err := io.Copy(fileWriter, inputFile); err != nil {
				errs <- fmt.Errorf("Error compressing file %s to zip: %w", file, err)
				return
			}

			zipWriter.Close()
		}()
	}
}
