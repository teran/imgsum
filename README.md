# Usage

The tool would produce 64-bit integer as a hash

```
Usage: ./bin/imgsum-darwin-amd64 [OPTION]... [FILE]...
Print or check image Average hashes
  -check
    read average hashes from the FILEs and check them
  -find-duplicates
    read average hashes from the FILEs and find duplicates

Examples:
  ./bin/imgsum-darwin-amd64 file.jpg
  ./bin/imgsum-darwin-amd64 file.jpg | tee /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -check /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -find-duplicates /tmp/database.txt
```

# NOTES

 * Image format supported at the moment: BMP, JPEG, GIF, PNG, CR2(Canon RAW), DNG, TIFF

# TODO

 * Add `-c` key working the same way as in *sum GNU utils
 * NEF format support
 * DNG format support
 * ORF format support
