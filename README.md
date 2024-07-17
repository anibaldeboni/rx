# rx

Rom eXpander is a cli tool designed to help you to manage your rom collection.
It can compress individual rom files into zip files and inflate content files from zip files.

# Usage

```sh
rx zip /my/rom/path --output /path/to/zip-files
```

```sh
rz unzip /my/zipped/roms --output /path/to/inflated/roms
```

If you don't set `--output` the files will be inflated/deflated to the same directory of the original files. Likewise if the dir path is not provided it current working dir will be used
