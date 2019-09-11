package image

import (
	"crypto/sha256"
	"encoding/hex"
	stdimage "image"
	"os"
	"regexp"

	"github.com/nf/cr2"
)

var (
	reCanon = regexp.MustCompile("(?i).cr(2)$")
)

// Image type
type Image struct {
	filename string
	fp       *os.File
	image    stdimage.Image
}

// NewImage creates new Image object
func NewImage(filename string) (*Image, error) {
	i := new(Image)
	i.filename = filename

	err := i.open()
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Image) open() error {
	var err error
	i.fp, err = os.Open(i.filename)
	if err != nil {
		return err
	}
	defer i.fp.Close()

	if reCanon.Match([]byte(i.filename)) {
		err = i.openCanonRaw()
	} else {
		err = i.openStdImage()
	}

	if err != nil {
		return err
	}
	return nil
}

func (i *Image) openStdImage() error {
	var err error
	i.image, _, err = stdimage.Decode(i.fp)
	if err != nil {
		return err
	}

	return nil
}

func (i *Image) openCanonRaw() error {
	var err error
	i.image, err = cr2.Decode(i.fp)
	if err != nil {
		return err
	}
	return nil
}

// Hexdigest produces image hash and returns it as string
// actually the string is SHA256 from the hasher's value
func (i *Image) Hexdigest(hashKind HashType, res int) (string, error) {
	hasher, err := getHasher(hashKind, res)
	dh, err := hasher.Hash(i.image)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(dh)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Filename returns image's filename
func (i *Image) Filename() string {
	return i.filename
}
