package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Soreil/arw"
	"github.com/brett-lempereur/ish"
	"github.com/nf/cr2"
)

type JsonOutput struct {
	Duplicates [][]string `json:"duplicates"`
	Count      int        `json:"count"`
}

var (
	re_canon = regexp.MustCompile("(?i).cr(2)$")
	re_sony  = regexp.MustCompile("(?i).(arw|sr2)$")
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

func deduplicate(filename string, json_output bool) error {
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

			if len(files[hash]) == 2 {
				counter = append(counter, hash)
			}
		}

		line, err = r.ReadString(10)
	}

	if json_output {
		out := JsonOutput{}
		for key := range counter {
			out.Duplicates = append(out.Duplicates, files[counter[key]])
		}
		out.Count = len(counter)
		jsonString, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonString))
	} else {
		for key := range counter {
			fmt.Printf("%v:\n", counter[key])
			for file := range files[counter[key]] {
				fmt.Println(files[counter[key]][file])
			}
			fmt.Println("")
		}
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
	} else if re_sony.Match([]byte(filename)) {
		header, err := arw.ParseHeader(fp)
		meta, err := arw.ExtractMetaData(fp, int64(header.Offset), 0)
		if err != nil {
			return nil, err
		}

		var jpegOffset uint32
		var jpegLength uint32
		for i := range meta.FIA {
			switch meta.FIA[i].Tag {
			case arw.JPEGInterchangeFormat:
				jpegOffset = meta.FIA[i].Offset
			case arw.JPEGInterchangeFormatLength:
				jpegLength = meta.FIA[i].Offset
			}
		}
		jpg, err := arw.ExtractThumbnail(fp, jpegOffset, jpegLength)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(jpg)
		img, err = jpeg.Decode(reader)
		if err != nil {
			return nil, err
		}
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
		fmt.Printf("    read average hashes from the FILEs and find duplicates\n")
		fmt.Printf("  -json-output\n")
		fmt.Printf("    Return duplicates as a JSON(useful for IPC)\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s file.jpg\n", os.Args[0])
		fmt.Printf("  %s file.jpg | tee /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -check /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -find-duplicates /tmp/database.txt\n", os.Args[0])
	}

	check_mode := flag.Bool("check", false, "")
	deduplicate_mode := flag.Bool("find-duplicates", false, "")
	json_output := flag.Bool("json-output", false, "")

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *check_mode == true {
		for file := range flag.Args() {
			err := checkFiles(flag.Arg(file))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	} else if *deduplicate_mode {
		for file := range flag.Args() {
			err := deduplicate(flag.Arg(file), *json_output)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	} else {
		for file := range flag.Args() {
			err := calculate(flag.Arg(file))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}
