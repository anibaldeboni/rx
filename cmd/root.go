package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/anibaldeboni/rx/files"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rx",
		Short: "rx is a cli tool to zip and unzip individual rom files.",
		Long: `Rom eXpander is a cli tool designed to help you to manage your rom collection.
It can compress individual rom files into zip files and inflate content files from zip files.
	`,
		Version: Version,
	}

	options       *Options
	walkRecursive bool
	outputDir     string
	workers       int
	Version       = "dev"
)

type Options struct {
	FindFunc files.FindFilesFunc
	Workers  int
	Output   string
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&walkRecursive, "recursive", "r", false, "walk recursively through directories")
	cwd, _ := os.Getwd()
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", cwd, "output directory")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", runtime.NumCPU(), "number of workers to use")
}

func FindFilesStrategy() files.FindFilesFunc {
	if walkRecursive {
		return files.FindRecursive
	}

	return files.FindFilesInRootDirectory
}

func HandleErrors(errs <-chan error) {
	go func(errs <-chan error) {
		for err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}(errs)
}

func validateCmdPath(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Path is required")
	}

	if err := checkPath(args[0]); err != nil {
		return fmt.Errorf("Could not run: %w", err)
	}

	if flag := cmd.Flag("output").Value.String(); flag != "" {
		options.Output = flag
	}

	return nil
}

func checkPath(path string) error {
	switch p, err := os.Stat(path); {
	case err != nil:
		return fmt.Errorf("Invalid path %s: %w", path, err)
	case p == nil:
		return fmt.Errorf("Path %s does not exist", path)
	case !p.IsDir():
		return fmt.Errorf("Path %s is not a directory", path)
	default:
		return nil
	}
}

func Execute() {
	options = &Options{
		FindFunc: FindFilesStrategy(),
		Workers:  workers,
		Output:   outputDir,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
