# Usage

The tool would produce 64-bit integer as a hash

```
Usage: ./bin/imgsum-darwin-amd64 [OPTION]... [FILE]...
Print or check image Average hashes
  -concurrency
    Amount of routines to spawn at the same time(8 by default for your system)
  -find-duplicates
    read average hashes from the FILEs and find duplicates
  -json-input
    Read file list from stdin as a JSON({'files':['file1', 'file2']}) and calculate their hash
  -json-output
    Return duplicates as a JSON(useful for IPC)
  -version
    Print imgsum version

Examples:
  ./bin/imgsum-darwin-amd64 file.jpg
  ./bin/imgsum-darwin-amd64 file.jpg | tee /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -check /tmp/database.txt
  ./bin/imgsum-darwin-amd64 -find-duplicates /tmp/database.txt
  cat input.json | ./bin/imgsum-darwin-amd64 -json-input
```

# JSON input

imgsum supports receiving file list to calculate hashes from STDIN which is useful
for IPC. The scheme for the JSON should be the following:

```
{
  "files": [
    "/Volumes/Disk1/file1.cr2",
    "/Volumes/Disk1/file2.jpg",
    "/Volumes/Disk1/file3.jpg"
  ]
}
```

Golang model:

```
type JsonInput struct {
	Files []string `json:"files"`
}
```

# NOTES

Image format supported and tested:
* Adobe Digital Negative(`*.dng`)
* Canon RAW(`*.cr2` - only, `*.crw` is not supported yet)
* Epson RAW(`*.erf`)
* Hasselblad 3FR(`*.3fr`)
* JPEG
* Kodak RAW(`*.kdc` - verified on Kodak DC50, DC120. Easyshare Z1015 RAW files doesn't work)
* Leaf RAW(`*.mos` - verified on Aptus 22, Aptus 75 doesn't work)
* Nikon RAW(`*.nef` - only, `*.nrw` is not supported yet)
* TIFF


# TODO

 * ORF format support
