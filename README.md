# Usage

The tool would produce 64-bit integer as a hash

```
Usage: ./bin/imgsum-darwin-amd64 [OPTION]... [FILE]...
Print or check image Average hashes
  -check
    read average hashes from the FILEs and check them
  -find-duplicates
    read average hashes from the FILEs and find duplicates
  -json-output
    Return duplicates as a JSON(useful for IPC)

Examples:
  ./bin/imgsum-darwin-amd64 file.jpg
  ./bin/imgsum-darwin-amd64 file.jpg | tee /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -check /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -find-duplicates /tmp/database.txt
```

# NOTES

Image format supported and tested:
* Adobe Digital Negative(`*.dng`)
* Canon RAW(`*.cr2` - only, `*.crw` is not supported yet)
* Epson RAW(`*.erf`)
* Hasselblad 3FR(`*.3fr`)
* JPEG
* Kodak RAW(`*.kdc` - verified on Kodak DC50, DC120. Easyshare Z1015 RAW files doesn't work)
* Nikon RAW(`*.nef` - only, `*.nrw` is not supported yet)
* TIFF
* Sony RAW(`*.arw`, `*.sr2`) - Experimental at the moment


# TODO

 * ORF format support
