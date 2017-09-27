// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

package imghash

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

// SearchResult is returned by Database.Find.
type SearchResult struct {
	Path     string // Image path, relative to Database.Root
	Hash     uint64 // Perceptual Image hash.
	Distance uint64 // Hamming Distance to search term.
}

// ResultSet holds search results, sortable by Hamming Distance.
type ResultSet []*SearchResult

func (r ResultSet) Len() int           { return len(r) }
func (r ResultSet) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ResultSet) Less(i, j int) bool { return r[i].Distance < r[j].Distance }

// Entry represents a single database entry.
type Entry struct {
	Path    string // Image path, relative to Database.Root
	Hash    uint64 // Perceptual Image hash.
	ModTime int64  // Last-Modified timestamp for this file.
}

// A Database holds a listing of Perceptual hashes, mapped
// to image file paths.
//
// Note: This is a very naive implementation that can benefit
// a great deal from optimization.
type Database struct {
	Root    string           // Database root path.
	entries []*Entry         // List of entries.
	pathMap map[string]int   //(private) map of file paths to Entry index
	hashMap map[uint64][]int //(private) map of file hashes to Entry indexes
}

// NewDatabase creates a new, empty database.
func NewDatabase() *Database {
	return &Database{pathMap: make(map[string]int), hashMap: make(map[uint64][]int)}
}

// Find finds all entries which have a Hamming Diance <= to the
// specified distance with the given hash.
// The list is sorted by relevance.
func (d *Database) Find(hash, distance uint64) ResultSet {
	var rs ResultSet
	var dist uint64

	//shortcut the enumeration and do a hash lookup if the distance is zero.
	if 0 == dist {
		for _, i := range d.hashMap[hash] {
			rs = append(rs, &SearchResult{
				Path:     d.entries[i].Path,
				Hash:     d.entries[i].Hash,
				Distance: 0,
			})
		}
	} else {
		for _, e := range d.entries {
			if nil == e {
				continue
			}
			dist = Distance(e.Hash, hash)

			if dist <= distance {
				rs = append(rs, &SearchResult{
					Path:     e.Path,
					Hash:     e.Hash,
					Distance: dist,
				})
			}
		}
	}

	sort.Sort(rs)
	return rs
}

// Load loads a database from the given file.
// Leave the filename empty to use the default file.
func (d *Database) Load(file string) (err error) {
	if len(file) == 0 {
		file = os.Getenv("IMGHASH_DB")
	}

	fd, err := os.Open(file)
	if err != nil {
		return
	}

	defer fd.Close()

	r := bufio.NewReader(fd)

	var line []byte
	var entry *Entry

	line, err = r.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		return
	}

	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return errors.New("Invalid database file.")
	}

	d.Root = string(line)

	for {
		line, err = r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		line = bytes.TrimSpace(line)
		if len(line) < 34 {
			continue
		}

		entry = new(Entry)
		entry.Path = string(line[33:])

		entry.Hash, err = strconv.ParseUint(string(line[:16]), 16, 64)
		if err != nil {
			return
		}

		entry.ModTime, err = strconv.ParseInt(string(line[18:32]), 16, 64)
		if err != nil {
			return
		}

		d.AddEntry(entry)
	}

	return
}

func (d *Database) AddEntry(entry *Entry) {
	d.entries = append(d.entries, entry)
	newIndex := len(d.entries) - 1
	d.pathMap[entry.Path] = newIndex
	d.hashMap[entry.Hash] = append(d.hashMap[entry.Hash], newIndex)
}

// Remove the entry without reshuffling the whole database.
// Note this means the array may have nil elements
func (d *Database) DeleteEntry(index int) {
	entry := d.entries[index]
	d.entries[index] = nil
	delete(d.pathMap, entry.Path)

	//there may be multiple entries with the same hash, so we rebuild the array
	for i, e := range d.hashMap[entry.Hash] {
		if e == index {
			d.hashMap[entry.Hash][i] = d.hashMap[entry.Hash][len(d.hashMap[entry.Hash])-1]
			d.hashMap[entry.Hash] = d.hashMap[entry.Hash][:len(d.hashMap[entry.Hash])-1]
			break
		}
	}

}

// Save saves the database to the given file.
// Leave the filename empty to use the default file.
func (d *Database) Save(file string) (err error) {
	if len(file) == 0 {
		file = os.Getenv("IMGHASH_DB")
	}

	fd, err := os.Create(file)
	if err != nil {
		return
	}

	defer fd.Close()

	fmt.Fprintf(fd, "%s\n", d.Root)

	for _, e := range d.entries {
		if nil != e {
			fmt.Fprintf(fd, "%016x %015x %s\n", e.Hash, e.ModTime, e.Path)
		}
	}

	return
}

// Set adds the given file if it doesn't already exist.
// Otherwise it overwrites the existing one.
func (d *Database) Set(file string, modtime int64, hash uint64) {
	index := d.IndexFile(file)

	if index == -1 {
		d.AddEntry(&Entry{file, hash, modtime})
		return
	}

	f := d.entries[index]
	f.ModTime = modtime
	f.Hash = hash
}

// IsNew returns true if the given file has been updated
// since it was last stored in the database.
func (d *Database) IsNew(file string, modtime int64) bool {
	index := d.IndexFile(file)

	if index == -1 {
		// Non-existant entry is always new.
		return true
	}

	return d.entries[index].ModTime != modtime
}

// IndexFile returns the index for the given file.
func (d *Database) IndexFile(file string) int {
	i, ok := d.pathMap[file]

	if !ok {
		return -1
	}
	return i
}

// IndexHash returns the indices for files with the given hash.
// There can be more than one of them.
func (d *Database) IndexHash(hash uint64) []int {
	return d.hashMap[hash]
}
