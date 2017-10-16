package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"imgsum/image"
)

type JsonOutput struct {
	Duplicates [][]string `json:"duplicates"`
	Count      int        `json:"count"`
}

func calculate(file string) (error) {
	i, err := image.NewImage(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, file, err.Error())
		return err
	}

	h, err := i.Hexdigest()
	if err != nil {
		fmt.Fprintln(os.Stderr, file, err.Error())
		return err
	}

	fmt.Printf("%v  %v\n", h, i.Filename())
	return nil
}

func check(checksumFile string) error {
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
			i, err := image.NewImage(fields[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, fields[1], err.Error())
				continue
			}

			h, err := i.Hexdigest()
			if err != nil {
				fmt.Fprintln(os.Stderr, fields[1], err.Error())
				continue
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
			check(flag.Arg(file))
		}
	} else if *deduplicate_mode {
		for file := range flag.Args() {
			deduplicate(flag.Arg(file), *json_output)
		}
	} else {
		for file := range flag.Args() {
			calculate(flag.Arg(file))
		}
	}
}
