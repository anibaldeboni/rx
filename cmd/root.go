package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/anibaldeboni/rx/files"
	"github.com/anibaldeboni/rx/styles"
	"github.com/spf13/cobra"
)

var Version = "dev"

const ZipExtension = ".zip"

type Cli struct {
	Workers   int
	Output    string
	Version   string
	Recursive bool
	errs      chan error
	wg        *sync.WaitGroup
}

type WorkerFunc func(<-chan string)

func (c *Cli) SetupWorkers(worker WorkerFunc, path string) {
	log.Printf("Looking for files at %s\n", styles.DarkPink(path))
	files := c.FindFiles()(path, c.errs)

	log.Printf("Using %s workers\n", styles.LightBlue(fmt.Sprintf("%d", c.Workers)))

	c.wg.Add(c.Workers)
	for i := 0; i < c.Workers; i++ {
		go worker(files)
	}

	c.wg.Wait()
	log.Println(styles.Green("Done!"))
}

func (c *Cli) FindFiles() files.FindFilesFunc {
	if c.Recursive {
		return files.FindRecursive
	}

	return files.FindFilesInRootDirectory
}

func HandleErrors(errs <-chan error) {
	go func(errs <-chan error) {
		for err := range errs {
			log.Println(err)
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

	return nil
}

func checkPath(path string) error {
	switch p, err := os.Stat(path); {
	case err != nil:
		return fmt.Errorf("Invalid path %s: %w", styles.DarkRed(path), err)
	case p == nil:
		return fmt.Errorf("Path %s does not exist", styles.DarkRed(path))
	case !p.IsDir():
		return fmt.Errorf("Path %s is not a directory", styles.DarkRed(path))
	default:
		return nil
	}
}

func Execute() {
	errs := make(chan error)
	defer close(errs)

	HandleErrors(errs)

	cli := &Cli{
		Version: Version,
		errs:    errs,
		wg:      &sync.WaitGroup{},
	}

	rootCmd := &cobra.Command{
		Use:   "rx",
		Short: "rx is a cli tool to zip and unzip individual rom files.",
		Long: `Rom eXpander is a cli tool designed to help you to manage your rom collection.
It can compress individual rom files into zip files and inflate content files from zip files.
	`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&cli.Recursive, "recursive", "r", false, "walk recursively through directories")
	cwd, _ := os.Getwd()
	rootCmd.PersistentFlags().StringVarP(&cli.Output, "output", "o", cwd, "output directory")
	rootCmd.PersistentFlags().IntVarP(&cli.Workers, "workers", "w", runtime.NumCPU(), "number of workers to use")

	rootCmd.AddCommand(cli.zipCmd())
	rootCmd.AddCommand(cli.unzipCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
