package main

import (
  "flag"
  "fmt"
  "image"
  _ "image/gif"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "regexp"
  "strconv"

  "github.com/jteeuwen/imghash"
  "github.com/nf/cr2"
)

var (
  re_canon = regexp.MustCompile(".cr(2|w)$")
)

func calculate(file string) (error) {
  var img image.Image
  img, err := getImage(file)
  if err != nil {
    fmt.Fprintln(os.Stderr, file, err.Error())
    return err
  }

  h, err := hash(img)
  if err != nil {
    fmt.Fprintln(os.Stderr, file, err.Error())
    return err
  }

  fmt.Printf("%v  %v\n", strconv.Itoa(int(h)), file)
  return nil
}

func checkFiles(checksum_file string) {}

func getImage(filename string) (image.Image, error) {
  var img image.Image
  var err error
  var fp *os.File

  fp, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
  defer fp.Close()

  if re_canon.Match([]byte(filename)) {
    img, err = cr2.Decode(fp)
  } else {
    img, _, err = image.Decode(fp)
  }

  if err != nil {
    return nil, err
  }

  return img, nil
}

func hash(img image.Image) (uint64, error) {
	return imghash.Average(img), nil
}

func main() {
  flag.Usage = func() {
    fmt.Printf("Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
    fmt.Printf("Print or check image Average hashes\n")
  }

  flag.Parse()
  if flag.NArg() < 1 {
    flag.Usage()
    os.Exit(1)
  }

  for file := range flag.Args() {
    calculate(flag.Arg(file))
  }
}
