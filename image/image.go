package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	stdimage "image"
	"image/jpeg"
	"os"
	"regexp"

	"github.com/Soreil/arw"
	"github.com/brett-lempereur/ish"
	"github.com/nf/cr2"
)

var (
	reCanon = regexp.MustCompile("(?i).cr(2)$")
	reSony  = regexp.MustCompile("(?i).(arw|sr2)$")
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
	} else if reSony.Match([]byte(i.filename)) {
		err = i.openSonyRaw()
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

func (i *Image) openSonyRaw() error {
	var jpegOffset uint32
	var jpegLength uint32

	header, err := arw.ParseHeader(i.fp)
	if err != nil {
		return err
	}

	meta, err := arw.ExtractMetaData(i.fp, int64(header.Offset), 0)
	if err != nil {
		return err
	}

	for i := range meta.FIA {
		switch meta.FIA[i].Tag {
		case arw.JPEGInterchangeFormat:
			jpegOffset = meta.FIA[i].Offset
		case arw.JPEGInterchangeFormatLength:
			jpegLength = meta.FIA[i].Offset
		}
	}

	jpg, err := arw.ExtractThumbnail(i.fp, jpegOffset, jpegLength)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jpg)
	i.image, err = jpeg.Decode(reader)
	if err != nil {
		return err
	}
	return nil
}

// Hexdigest produces image hash and returns it as string
// actually the string is SHA256 from the hasher's value
func (i *Image) Hexdigest() (string, error) {
	hasher := ish.NewAverageHash(1024, 1024)
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
