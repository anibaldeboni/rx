package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anibaldeboni/rx/styles"
	"github.com/spf13/cobra"
)

func (c *Cli) zipCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "zip [path]",
		Short:   "compress rom files.",
		Long:    `Compress rom files into individual zip files.`,
		PreRunE: c.validatePath,
		Run:     c.zip,
	}
}

func (c *Cli) zip(cmd *cobra.Command, args []string) {
	c.SetupWorkers(c.zipWorker, c.Cwd)
}

func (c *Cli) zipWorker(files <-chan string, addr int) {
	defer c.wg.Done()

	for file := range files {
		func() {
			zipFile := filepath.Base(file)
			outputFileName := strings.TrimSuffix(zipFile, filepath.Ext(file)) + ZipExtension
			outputFilePath := filepath.Join(c.Output, outputFileName)

			outputFile, err := os.Create(outputFilePath)
			defer outputFile.Close()
			if err != nil {
				c.errs <- fmt.Errorf("[%s] Error creating file %s: %w", fmtWorker(addr), outputFileName, err)
				return
			}

			inputFile, err := os.Open(file)
			defer inputFile.Close()
			if err != nil {
				c.errs <- fmt.Errorf("[%s] Error opening file %s: %w", fmtWorker(addr), file, err)
				return
			}

			zipWriter := zip.NewWriter(outputFile)

			fileWriter, err := zipWriter.Create(zipFile)
			if err != nil {
				c.errs <- fmt.Errorf("[%s] Error creating file %s in zip: %w", fmtWorker(addr), file, err)
				return
			}

			log.Printf("[%s] Compressing %s\n", fmtWorker(addr), styles.LightBlue(zipFile))

			if _, err := io.Copy(fileWriter, inputFile); err != nil {
				c.errs <- fmt.Errorf("[%s] Error compressing file %s to zip: %w", fmtWorker(addr), file, err)
				return
			}

			zipWriter.Close()
		}()
	}
}
