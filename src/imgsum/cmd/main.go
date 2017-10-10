package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/brett-lempereur/ish"
	"github.com/nf/cr2"
)

var (
	re_canon = regexp.MustCompile(".cr(2|w)$")
	re_tiff  = regexp.MustCompile(".tiff?$")
	re_bmp   = regexp.MustCompile(".bmp$")
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

	fmt.Printf("%v  %v\n", h, file)
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

			if fields[0] == h {
				fmt.Printf("%v: OK\n", fields[1])
			} else {
				fmt.Printf("%v: FAILED\n", fields[1])
			}
		}

		line, err = r.ReadString(10)
	}
	return nil
}

func deduplicate(filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, filename, err.Error())
		return err
	}
	defer fp.Close()

	files := make(map[string][]string)
	var counter []string

	r := bufio.NewReader(fp)
	line, err := r.ReadString(10)
	for err != io.EOF {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			hash := fields[0]
			file := fields[1]

			files[hash] = append(files[hash], file)

			if len(files[hash]) > 1 {
				counter = append(counter, hash)
			}
		}

		line, err = r.ReadString(10)
	}

	for key := range counter {
		fmt.Printf("%v:\n", counter[key])
		for file := range files[counter[key]] {
			fmt.Println(files[counter[key]][file])
		}
		fmt.Println("")
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
	} else if re_bmp.Match([]byte(filename)) {
		img, err = bmp.Decode(fp)
	} else {
		img, _, err = image.Decode(fp)
	}

	if err != nil {
		return nil, err
	}

	return img, nil
}

func hash(img image.Image) (string, error) {
	hasher := ish.NewAverageHash(1024, 1024)
	dh, err := hasher.Hash(img)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(dh)

	return hex.EncodeToString(h.Sum(nil)), nil
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
		fmt.Printf("Print or check image Average hashes\n")
		fmt.Printf("  -check\n")
		fmt.Printf("    read average hashes from the FILEs and check them\n")
		fmt.Printf("  -find-duplicates\n")
		fmt.Printf("    read average hashes from the FILEs and find duplicates\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s file.jpg\n", os.Args[0])
		fmt.Printf("  %s file.jpg | tee /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -check /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -find-duplicates /tmp/database.txt\n", os.Args[0])
	}

	check := flag.Bool("check", false, "")
	dedup := flag.Bool("find-duplicates", false, "")

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *check == true {
		for file := range flag.Args() {
			err := checkFiles(flag.Arg(file))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	} else if *dedup {
		for file := range flag.Args() {
			err := deduplicate(flag.Arg(file))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	} else {
		for file := range flag.Args() {
			err := calculate(flag.Arg(file))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
}
