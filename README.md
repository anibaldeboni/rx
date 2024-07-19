# rx

Rom eXpander is a cli tool designed to help you manage your rom collection.
It can compress individual rom files into zip files and inflate content files from zip files.

# Usage

```sh
  rx [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  unzip       inflate rom files.
  zip         compress rom files.

Flags:
  -h, --help            help for rx
  -o, --output string   output directory
  -r, --recursive       walk recursively through directories
  -v, --version         version for rx
  -w, --workers int     number of workers to use (default 8)
```

If you don't set `--output` the files will be inflated/deflated to the same directory of the original files. Likewise if the dir path is not provided the current working dir will be used.
