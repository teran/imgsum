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

func hash(img image.Image) (uint64, error) {
	return imghash.Average(img), nil
}

func getImage(filename string) (image.Image, error) {
  var img image.Image
  var err error
  var fp *os.File

  fp, err = os.Open(filename)
	if err != nil {
    fmt.Printf("Error opening file:", err)
		os.Exit(1)
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

func main() {
  var img image.Image

  flag.Usage = func() {
    fmt.Printf("Usage: %s <image>\n", os.Args[0])
  }
  flag.Parse()
  if flag.NArg() < 1 {
    flag.Usage()
    os.Exit(1)
  }

  file := flag.Arg(0)

  img, err := getImage(file)
	if err != nil {
    fmt.Println(err)
		os.Exit(1)
	}

  h, err := hash(img)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Printf("%v  %v\n", strconv.Itoa(int(h)), file)
}
