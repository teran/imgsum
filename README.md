# Usage

```
Usage: ./bin/imgsum-darwin-amd64 <image>
```

The tool would produce 64-bit integer as a hash


Example:

```
$ ./bin/imgsum-darwin-amd64 /Volumes/Backpack/Data/Photos/Lightroom/2015/08/29/20150829-00334-5113.cr2
220902682783518  /Volumes/Backpack/Data/Photos/Lightroom/2015/08/29/20150829-00334-5113.cr2
```

# NOTES

 * Image format supported at the moment: JPEG, GIF, PNG, CR2(Canon RAW)

# TODO

 * Add `-c` key working the same way as in *sum GNU utils
 * NEF format support
 * DNG format support
 * ORF format support
