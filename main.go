package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"

	"imgsum/image"
)

type JsonOutput struct {
	Duplicates [][]string `json:"duplicates"`
	Count      int        `json:"count"`
}

type JsonInput struct {
	Files []string `json:"files"`
}

var wg sync.WaitGroup
var Version = "No version specified(probably trunk build)"

func calculate(file string) error {
	i, err := image.NewImage(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, file, err.Error())
		wg.Done()
		return err
	}

	h, err := i.Hexdigest()
	if err != nil {
		fmt.Fprintln(os.Stderr, file, err.Error())
		wg.Done()
		return err
	}

	fmt.Printf("%v  %v\n", h, i.Filename())
	wg.Done()
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
		fields := strings.SplitN(line, " ", 2)
		if len(fields) == 2 {
			hash := strings.TrimSpace(fields[0])
			file := strings.TrimSpace(fields[1])

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
		fmt.Printf("  -concurrency %v\n", runtime.NumCPU())
		fmt.Printf("    Amount of routines to spawn at the same time(%v by default for your system)\n", runtime.NumCPU())
		fmt.Printf("  -find-duplicates\n")
		fmt.Printf("    read average hashes from the FILEs and find duplicates\n")
		fmt.Printf("  -json-input\n")
		fmt.Printf("    Read file list from stdin as a JSON({'files':['file1', 'file2']}) and calculate their hash\n")
		fmt.Printf("  -json-output\n")
		fmt.Printf("    Return duplicates as a JSON(useful for IPC)\n")
		fmt.Printf("  -version\n")
		fmt.Printf("    Print imgsum version\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  %s file.jpg\n", os.Args[0])
		fmt.Printf("  %s file.jpg | tee /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -check /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  %s -find-duplicates /tmp/database.txt\n", os.Args[0])
		fmt.Printf("  cat input.json | %s -json-input\n", os.Args[0])
	}

	concurrency := flag.Int("concurrency", runtime.NumCPU(), "")
	deduplicate_mode := flag.Bool("find-duplicates", false, "")
	json_input := flag.Bool("json-input", false, "")
	json_output := flag.Bool("json-output", false, "")
	version := flag.Bool("version", false, "")

	flag.Parse()
	if flag.NArg() < 1 && !*json_input && !*version {
		flag.Usage()
		os.Exit(1)
	}

	if *deduplicate_mode {
		for file := range flag.Args() {
			deduplicate(flag.Arg(file), *json_output)
		}
	} else if *version == true {
		fmt.Println(Version)
	} else {
		var files []string
		if *json_input {
			var jsonInput JsonInput
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(data, &jsonInput); err != nil {
				panic(err)
			}

			files = jsonInput.Files
		} else {
			files = flag.Args()
		}

		sem := make(chan bool, *concurrency)
		for file := range files {
			sem <- true
			filename := files[file]
			wg.Add(1)
			go func() {
				calculate(filename)
				defer func() {
					<-sem
				}()
			}()
		}

		for i := 0; i < cap(sem); i++ {
			sem <- true
		}
		wg.Wait()
	}
}
