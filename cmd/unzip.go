package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rx/styles"
	"github.com/spf13/cobra"
)

func (c *Cli) unzipCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "unzip [path]",
		Short:   "inflate rom files.",
		Long:    `Inflate content files from zip files`,
		PreRunE: c.validatePath,
		Run:     c.unzip,
	}
}

func (c *Cli) unzip(cmd *cobra.Command, args []string) {
	c.SetupWorkers(c.unzipWorker, c.Cwd)
}

func (c *Cli) unzipWorker(files <-chan string, addr int) {
	defer c.wg.Done()

	for file := range files {
		if filepath.Ext(file) != ZipExtension {
			continue
		}

		zipFile, err := zip.OpenReader(file)
		if err != nil {
			c.errs <- fmt.Errorf("[%s] Error opening zip archive %s: %w", fmtWorker(addr), styles.DarkRed(file), err)
		}
		defer zipFile.Close()
		for _, zipContent := range zipFile.File {
			outputFilePath := filepath.Join(c.Output, zipContent.Name)

			log.Printf("[%s] Inflating %s\n", fmtWorker(addr), styles.LightBlue(zipContent.Name))

			if zipContent.FileInfo().IsDir() {
				if err := os.MkdirAll(outputFilePath, os.ModePerm); err != nil {
					c.errs <- fmt.Errorf("[%s] Error creating directory %s: %w", fmtWorker(addr), styles.DarkRed(outputFilePath), err)
				}
				continue
			}

			if err := os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm); err != nil {
				c.errs <- fmt.Errorf("[%s] Error creating directory %s: %w", fmtWorker(addr), styles.DarkRed(filepath.Dir(outputFilePath)), err)
			}

			dstFile, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipContent.Mode())
			if err != nil {
				c.errs <- fmt.Errorf("[%s] Error creating file %s: %w", fmtWorker(addr), styles.DarkRed(outputFilePath), err)
			}

			srcFile, err := zipContent.Open()
			if err != nil {
				c.errs <- fmt.Errorf("[%s] Error compressed file %s: %w", fmtWorker(addr), styles.DarkRed(zipContent.Name), err)
			}
			if _, err := io.Copy(dstFile, srcFile); err != nil {
				c.errs <- fmt.Errorf("[%s] Error inflating file %s: %w", fmtWorker(addr), styles.DarkRed(zipContent.Name), err)
			}

			dstFile.Close()
			srcFile.Close()
		}
	}
}
