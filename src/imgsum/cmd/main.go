package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/tiff"

	"github.com/jteeuwen/imghash"
	"github.com/nf/cr2"
)

var (
	re_canon = regexp.MustCompile(".cr(2|w)$")
	re_tiff = regexp.MustCompile(".tiff?$")
)

func calculate(file string) error {
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

func checkFiles(checksumFile string) error {
	fp, err := os.Open(checksumFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, checksumFile, err.Error())
		return err
	}
	defer fp.Close()

	r := bufio.NewReader(fp)
	line, err := r.ReadString(10)
	for err != io.EOF {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			img, err := getImage(fields[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, fields[1], err.Error())
				return err
			}

			h, err := hash(img)
			if err != nil {
				fmt.Fprintln(os.Stderr, fields[1], err.Error())
				return err
			}

			if fields[0] == strconv.Itoa(int(h)) {
				fmt.Printf("%v: OK\n", fields[1])
			} else {
				fmt.Printf("%v: FAILED\n", fields[1])
			}
		}

		line, err = r.ReadString(10)
	}
	return nil
}

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
	} else if re_tiff.Match([]byte(filename)) {
		img, err = tiff.Decode(fp)
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
		fmt.Printf("  -check\n")
		fmt.Printf("    read average hashes from the FILEs and check them\n")
	}

	check := flag.Bool("check", false, "")

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *check == true {
		for file := range flag.Args() {
			checkFiles(flag.Arg(file))
		}
	} else {
		for file := range flag.Args() {
			calculate(flag.Arg(file))
		}
	}
}
