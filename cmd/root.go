package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/anibaldeboni/rx/files"
	"github.com/anibaldeboni/rx/styles"
	"github.com/spf13/cobra"
)

var Version = "dev"

const (
	ZipExtension        = ".zip"
	ErrPathNotExists    = "Path %s does not exist"
	ErrPathNotDirectory = "Path %s is not a directory"
	ErrInvalidPath      = "Invalid path %s: %w"
)

type Cli struct {
	Workers   int
	Output    string
	Version   string
	Cwd       string
	Recursive bool
	errs      chan error
	wg        *sync.WaitGroup
}

type WorkerFunc func(<-chan string, int)

func (c *Cli) SetupWorkers(worker WorkerFunc, path string) {
	start := time.Now()
	log.Printf("Looking for files at %s\n", styles.DarkPink(path))
	files := c.FindFiles()(path, c.errs)

	log.Printf("Using %s workers\n", styles.LightBlue(fmt.Sprintf("%d", c.Workers)))

	c.wg.Add(c.Workers)
	for i := 0; i < c.Workers; i++ {
		go worker(files, i+1)
	}

	c.wg.Wait()
	log.Println(styles.Green("Done!"), "Elapsed time:", time.Since(start))
}

func (c *Cli) FindFiles() files.FindFilesFunc {
	if c.Recursive {
		return files.FindRecursive
	}

	return files.FindFilesInRootDirectory
}

func fmtWorker(addr int) string {
	return styles.LightGreen(fmt.Sprintf("🛠️ %03d", addr))
}

func HandleErrors(errs <-chan error) {
	go func(errs <-chan error) {
		for err := range errs {
			log.Println(err)
		}
	}(errs)
}

func (c *Cli) validatePath(cmd *cobra.Command, args []string) error {
	var path string

	if len(args) < 1 {
		path = c.Cwd
	} else {
		path = args[0]
	}

	if err := checkPath(path); err != nil {
		return fmt.Errorf("Could not run: %w", err)
	}

	if _, err := os.Stat(c.Output); os.IsNotExist(err) {
		os.MkdirAll(c.Output, os.ModePerm)
	}

	log.Printf("Output directory: %s\n", styles.Green(c.Output))
	return nil
}

func checkPath(path string) error {
	switch p, err := os.Stat(path); {
	case os.IsNotExist(err):
		return fmt.Errorf("Path %s does not exist", styles.DarkRed(path))
	case !p.IsDir():
		return fmt.Errorf("Path %s is not a directory", styles.DarkRed(path))
	case err != nil:
		return fmt.Errorf("Invalid path %s: %w", styles.DarkRed(path), err)
	default:
		return nil
	}
}

func Execute() {
	errs := make(chan error)
	defer close(errs)

	log.SetFlags(log.Lmicroseconds)
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
	cli.Cwd = cwd
	rootCmd.PersistentFlags().StringVarP(&cli.Output, "output", "o", cwd, "output directory")
	rootCmd.PersistentFlags().IntVarP(&cli.Workers, "workers", "w", runtime.NumCPU(), "number of workers to use")

	rootCmd.AddCommand(cli.zipCmd())
	rootCmd.AddCommand(cli.unzipCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
