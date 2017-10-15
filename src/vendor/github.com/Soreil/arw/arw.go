//package arw implements basic support for Exif 2.3 according to CIPA DC-008-2012
package arw

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"errors"
	"math"
	"log"
)

//CIPA DC-008-2012 Table 1
type TIFFHeader struct {
	ByteOrder uint16
	FortyTwo  uint16
	Offset    uint32
}

//CIPA DC-008-2012 Chapter 4.6.2
type EXIFIFD struct {
	Count   uint16
	FIA     []IFDFIA
	FIAvals []FIAval
	Offset  uint32
}

func (e EXIFIFD) String() string {
	var result []string
	result = append(result, fmt.Sprintf("Count: %v", e.Count))
	for i := range e.FIA {
		var val string
		val = e.FIAvals[i].String()
		if int(e.FIA[i].Count)*e.FIA[i].Type.Len() <= 4 {
			switch e.FIA[i].Type {
			case BYTE:
				val = fmt.Sprintf("%x %x %x %x",(e.FIA[i].Offset>>24)&0xff,(e.FIA[i].Offset>>16)&0xff,(e.FIA[i].Offset>>8)&0xff,(e.FIA[i].Offset>>0)&0xff)
			case ASCII:
				val = fmt.Sprintf("%c %c %c %c",(e.FIA[i].Offset>>24)&0xff,(e.FIA[i].Offset>>16)&0xff,(e.FIA[i].Offset>>8)&0xff,(e.FIA[i].Offset>>0)&0xff)
			case SHORT:
				val = fmt.Sprintf("%v %v",(e.FIA[i].Offset>>16)&0xffff,(e.FIA[i].Offset>>0)&0xffff)
			case LONG:
				val = fmt.Sprint(e.FIA[i].Offset)
			case UNDEFINED:
				val = fmt.Sprintf("%x %x %x %x",(e.FIA[i].Offset>>24)&0xff,(e.FIA[i].Offset>>16)&0xff,(e.FIA[i].Offset>>8)&0xff,(e.FIA[i].Offset>>0)&0xff)
			case SSHORT:
				val = fmt.Sprintf("%v %v",int16((e.FIA[i].Offset>>16)&0xffff),int16((e.FIA[i].Offset>>0)&0xffff))
			case SLONG:
				val = fmt.Sprint(int32(e.FIA[i].Offset))
			}
		}
		result = append(result, fmt.Sprintf("%v: %v", e.FIA[i].Tag, val))
	}
	result = append(result, fmt.Sprintf("Offset to next EXIFIFD: %v", e.Offset))
	return strings.Join(result, "\n")
}

//IFD Field Interoperability Array
//CIPA DC-008-2012 Chapter 4.6.2
type IFDFIA struct {
	Tag    IFDtag
	Type   IFDtype
	Count  uint32
	Offset uint32
}

//CIPA DC-008-2012 Chapter 4.6.2
type FIAval struct {
	IFDtype
	ascii     *[]byte
	short     *[]uint16
	long      *[]uint32
	slong     *[]int32
	rat       *[]float32
}

type ShotInfoTags struct {
	_ [2]byte
	FaceInfoOffset uint16
	SonyDateTime [20]byte
	SonyImageHeight uint16
	SonyImageWidth uint16
	FacesDetected uint16
	FaceInfoLength uint16
	MetaVersion [16]byte
	_ [4]byte
	FaceInfo1
	FaceInfo2
}

type FaceInfo1 struct {
	Face1Position [4]uint16
	_ [24]byte
	Face2Position [4]uint16
	_ [24]byte
	Face3Position [4]uint16
	_ [24]byte
	Face4Position [4]uint16
	_ [24]byte
	Face5Position [4]uint16
	_ [24]byte
	Face6Position [4]uint16
	_ [24]byte
	Face7Position [4]uint16
	_ [24]byte
	Face8Position [4]uint16
	_ [24]byte
}

type FaceInfo2 struct {
	Face1Position [4]uint16
	_ [29]byte
	Face2Position [4]uint16
	_ [29]byte
	Face3Position [4]uint16
	_ [29]byte
	Face4Position [4]uint16
	_ [29]byte
	Face5Position [4]uint16
	_ [29]byte
	Face6Position [4]uint16
	_ [29]byte
	Face7Position [4]uint16
	_ [29]byte
	Face8Position [4]uint16
	_ [29]byte
}

func (f FIAval) String() string {
	var val string
	switch f.IFDtype {
	case 1,7:
		val = fmt.Sprint(*f.ascii)
	case 2:
		val = fmt.Sprint(string(*f.ascii))
	case 3:
		parts := make([]string,len(*f.short))
		for i, short := range *f.short {
			parts[i] = fmt.Sprint(short)
		}
		val = strings.Join(parts,", ")
	case 4:
		parts := make([]string,len(*f.long))
		for i, long := range *f.long {
			parts[i] = fmt.Sprint(long)
		}
		val = strings.Join(parts,", ")
	case 9:
		parts := make([]string,len(*f.slong))
		for i, slong := range *f.slong {
			parts[i] = fmt.Sprint(slong)
		}
		val = strings.Join(parts,", ")
	case 5:
		parts := make([]string,len(*f.rat))
		for i, rat := range *f.rat {
			parts[i] = fmt.Sprint(rat)
		}
		val = strings.Join(parts,", ")
	case 10:
		parts := make([]string,len(*f.rat))
		for i, rat := range *f.rat {
			parts[i] = fmt.Sprint(rat)
		}
		val = strings.Join(parts,", ")
	}

	return val
}

//go:generate stringer -type=IFDtag
type IFDtag uint16

//IFDtags mapping taken from http://www.exiv2.org/tags.html
const (
	NewSubFileType   IFDtag = 254
	ImageWidth IFDtag = 256
	ImageHeight IFDtag = 257
	BitsPerSample IFDtag = 258
	Compression      IFDtag = 259
	PhotometricInterpretation IFDtag = 262
	ImageDescription IFDtag = 270
	Make             IFDtag = 271
	Model            IFDtag = 272
	StripOffsets     IFDtag = 273
	Orientation      IFDtag = 274
	SamplesPerPixel IFDtag = 277
	RowsPerStrip IFDtag = 278
	StripByteCounts  IFDtag = 279
	XResolution      IFDtag = 282
	YResolution      IFDtag = 283
	PlanarConfiguration IFDtag = 284
	ResolutionUnit   IFDtag = 296
	Software         IFDtag = 305
	DateTime         IFDtag = 306
	SubIFDs          IFDtag = 330

	JPEGInterchangeFormat       IFDtag = 513
	JPEGInterchangeFormatLength IFDtag = 514
	YCbCrPositioning            IFDtag = 531

	ShotInfo IFDtag = 0x3000
	FileFormat IFDtag = 0xb000
	SonyModelID IFDtag =  0xb001
	CreativeStyle IFDtag = 0xb020
	LensSpec IFDtag = 0xb02a
	FullImageSize IFDtag = 0xb02b
	PreviewImageSize IFDtag = 0xb02c
	Tag9400 IFDtag = 0x9400 //Tag9400A-C

	//Following tags have been scavenged from the internet, most likely to do with the Sony raw data in ARW
	SonyRawFileType IFDtag = 0x7000
	SonyCurve IFDtag = 0x7010

	SR2SubIFDOffset IFDtag = 0x7200
	SR2SubIFDLength IFDtag = 0x7201
	SR2SubIFDKey IFDtag = 0x7221

	IDC_IFD IFDtag = 0x7240
	IDC2_IFD IFDtag = 0x7241
	MRWInfo IFDtag = 0x7250

	ExifTag             IFDtag = 34665
	GPSTag              IFDtag = 34853
	InteroperabilityTag IFDtag = 40965
	PrintImageMatching  IFDtag = 50341
	DNGPrivateData      IFDtag = 50740

	ExposureTime             IFDtag = 33434
	FNumber                  IFDtag = 33437
	ExposureProgram          IFDtag = 34850
	SpectralSensitivity      IFDtag = 34852
	ISOSpeedRatings          IFDtag = 34855
	OECF                     IFDtag = 34856
	SensitivityType IFDtag = 34864
	RecommendedExposureIndex IFDtag = 34866
	ExifVersion              IFDtag = 36864
	DateTimeOriginal         IFDtag = 36867
	DateTimeDigitized        IFDtag = 36868
	ComponentsConfiguration  IFDtag = 37121
	CompressedBitsPerPixel   IFDtag = 37122
	ShutterSpeedValue        IFDtag = 37377
	ApertureValue            IFDtag = 37378
	BrightnessValue          IFDtag = 37379
	ExposureBiasValue        IFDtag = 37380
	MaxApertureValue         IFDtag = 37381
	SubjectDistance          IFDtag = 37382
	MeteringMode             IFDtag = 37383
	LightSource              IFDtag = 37384
	Flash                    IFDtag = 37385
	FocalLength              IFDtag = 37386
	SubjectArea              IFDtag = 37396
	MakerNote                IFDtag = 37500
	UserComment              IFDtag = 37510
	SubsecTime               IFDtag = 37520
	SubsecTimeOriginal       IFDtag = 37521
	SubsecTimeDigitized      IFDtag = 37522
	FlashpixVersion          IFDtag = 40960
	ColorSpace               IFDtag = 40961
	PixelXDimension          IFDtag = 40962
	PixelYDimension          IFDtag = 40963
	RelatedSoundFile         IFDtag = 40964
	FlashEnergy              IFDtag = 41483
	SpatialFrequencyResponse IFDtag = 41484
	FocalPlaneXResolution    IFDtag = 41486
	FocalPlaneYResolution    IFDtag = 41487
	FocalPlaneResolutionUnit IFDtag = 41488
	SubjectLocation          IFDtag = 41492
	ExposureIndex            IFDtag = 41493
	SensingMethod            IFDtag = 41495
	FileSource               IFDtag = 41728
	SceneType                IFDtag = 41729
	CFAPattern               IFDtag = 41730
	CustomRendered           IFDtag = 41985
	ExposureMode             IFDtag = 41986
	WhiteBalance             IFDtag = 41987
	DigitalZoomRatio         IFDtag = 41988
	FocalLengthIn35mmFilm    IFDtag = 41989
	SceneCaptureType         IFDtag = 41990
	GainControl              IFDtag = 41991
	Contrast                 IFDtag = 41992
	Saturation               IFDtag = 41993
	Sharpness                IFDtag = 41994
	DeviceSettingDescription IFDtag = 41995
	SubjectDistanceRange     IFDtag = 41996
	ImageUniqueID            IFDtag = 42016
	LensSpecification        IFDtag = 42034
	LensModel                IFDtag = 42036

	CFARepeatPatternDim      IFDtag = 0x828d
	CFAPattern2              IFDtag = 0x828e

)

//IFD datatype, most datatypes translate in to C datatypes.
//go:generate stringer -type=IFDtype
type IFDtype uint16

const (
	UNKNOWNTYPE IFDtype = iota
	BYTE
	ASCII
	SHORT
	LONG
	RATIONAL
	_
	UNDEFINED
	SSHORT
	SLONG
	SRRATIONAL
)

//IFDType length in bytes
func (i IFDtype) Len() int {
	switch i {
	case BYTE, ASCII, UNDEFINED:
		return 1
	case SHORT, SSHORT:
		return 2
	case LONG, SLONG:
		return 4
	case RATIONAL, SRRATIONAL:
		return 8
	default:
		return -1
	}
}

//Anyone who thinks I'm switching byte order mid program is sorely mistaken.
var b binary.ByteOrder = binary.LittleEndian

//Parses a TIFF header to determine first IFD and endianness.
func ParseHeader(r io.ReadSeeker) (TIFFHeader, error) {
	endian := make([]byte, 2)
	r.Read(endian)
	switch string(endian) {
	case "II":
		b = binary.LittleEndian
	case "MM":
		b = binary.BigEndian
	default:
		return TIFFHeader{}, errors.New("failed to determine endianness: "+fmt.Sprint(endian))
	}
	r.Seek(-2, 1)

	var header TIFFHeader
	binary.Read(r, b, &header)
	if header.FortyTwo != 42 {
		return header,errors.New("found an endianness marker but no fixed 42, offset might be unrealiable")
	}
	return header, nil
}

//ExtractMetadata will return the first IFD from a TIFF document.
func ExtractMetaData(r io.ReadSeeker, offset int64, whence int) (meta EXIFIFD, err error) {
	r.Seek(offset, whence)
	binary.Read(r, b, &meta.Count)
	meta.FIA = make([]IFDFIA, int(meta.Count))
	binary.Read(r, b, &meta.FIA)
	binary.Read(r, b, &meta.Offset)

	meta.FIAvals = make([]FIAval, len(meta.FIA))
	for n, interop := range meta.FIA {
		meta.FIAvals[n].IFDtype = interop.Type

		//Offset field is actually the value
		if uint32(interop.Type.Len())*interop.Count <= 4 {
			switch interop.Type {
			case 1, 2, 7:
				values := make([]byte, interop.Count)
				for i := range values {
					values[i] = byte(((interop.Offset << uint32(8*i)) & 0xff000000) >> 24)
				}
				meta.FIAvals[n].ascii = &values
			case 3:
				values := make([]uint16, interop.Count)
				for i := range values {
					values[i] = uint16(((interop.Offset << uint32(16*i)) & 0xffff0000) >> 16)
				}
				meta.FIAvals[n].short = &values
			case 4:
				values := []uint32{interop.Count}
				meta.FIAvals[n].long = &values
			case 9:
				values := []int32{int32(interop.Count)}
				meta.FIAvals[n].slong = &values
			}
		} else {
			r.Seek(int64(interop.Offset), 0)
			switch interop.Type {
			case 1, 2, 7:
				values := make([]byte, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].ascii = &values
			case 3:
				values := make([]uint16, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].short = &values
			case 4:
				values := make([]uint32, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].long = &values
			case 9:
				values := make([]int32, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].slong = &values
			case 5:
				values := make([]uint32, interop.Count*2)
				binary.Read(r, b, &values)
				floats := make([]float32,interop.Count)
				for i := range floats {
					floats[i] = float32(values[i*2])/float32(values[(i*2)+1])
				}
				meta.FIAvals[n].rat = &floats
			case 10:
				values := make([]int32, interop.Count*2)
				binary.Read(r, b, &values)
				floats := make([]float32,interop.Count)
				for i := range floats {
					floats[i] = float32(values[i*2])/float32(values[(i*2)+1])
				}
				meta.FIAvals[n].rat = &floats
			}
		}
	}

	return
}

func DecryptSR2(r io.ReaderAt, offset uint32, length uint32,key [4]byte) ([]byte, error){
	buf := make([]byte,length)
	r.ReadAt(buf,int64(offset))

	var pad [128]byte

	for i := 0; i < 4; i++ {
		pad[i]=key[i]
	}
	pad[3] = pad[3] << 1 | (pad[0]^pad[2]) >> 31

	for i:=4; i < 127; i++ {
		pad[i] = (pad[i-4]^pad[i-2]) << 1 | (pad[i-3]^pad[i-1]) >> 31
	}

	for i := 127;i < int(length)+127; i++ {
		or :=  pad[(i+1) & 127] ^ pad[(i+65) & 127]
		pad[i & 127] = or
		buf[i-127] ^= byte(or)
	}

	return buf,nil
}

//ExtractThumbnail extracts an embedded JPEG thumbnail.
func ExtractThumbnail(r io.ReaderAt, offset uint32, length uint32) ([]byte, error) {
	jpegData := make([]byte, length)
	_, err := r.ReadAt(jpegData, int64(offset))
	if err != nil {
		return nil, err
	}
	return jpegData, nil
}

type pixelblock struct {
	max    uint16
	min    uint16
	maxidx uint8
	minidx uint8
	pix    [14]uint8
}

const pixelblocksize = 16

type pixel uint16

func (p pixelblock) String() string {
	return fmt.Sprintf("%011b\n%011b\n%04b\n%04b\n%08b", p.max, p.min, p.maxidx, p.minidx, p.pix)
}

func (p pixelblock) Decompress() [pixelblocksize]pixel {
	var pix [pixelblocksize]pixel
	factor := uint8(1 << uint8(math.Ceil(math.Log2(float64(p.max-p.min)/128))))
	var ordinary int
	for i := 0; i < pixelblocksize; i++ {
		switch i {
		case int(p.maxidx):
			pix[i] = pixel(p.max)
		case int(p.minidx):
			pix[i] = pixel(p.min)
		default:
			log.Println(p.min,p.max,p.max-p.min,factor)
			pix[i]=pixel(p.min) + pixel(p.pix[ordinary]*factor)
			ordinary++
			}
	}
	return pix
}

func readblock(s []byte) pixelblock {
	var p pixelblock

	p.max = ((uint16(s[0]) & 0xff) << 3) + (uint16(s[1])&0xe0)>>5
	p.min = ((uint16(s[1]) & 0x1f) << 6) + (uint16(s[2])&0xfc)>>2

	p.maxidx = ((s[2] & 0x03) << 2) + (s[3]&0xc0)>>6
	p.minidx = (s[3] & 0x3c) >> 2

	p.pix[0] = ((s[3] & 0x03) << 5) + ((s[4] & 0xf8) >> 3)
	p.pix[1] = ((s[4] & 0x07) << 4) + ((s[5] & 0xf0) >> 4)
	p.pix[2] = ((s[5] & 0x0f) << 3) + ((s[6] & 0xe0) >> 5)
	p.pix[3] = ((s[6] & 0x1f) << 2) + ((s[7] & 0xc0) >> 6)
	p.pix[4] = ((s[7] & 0x3f) << 1) + ((s[8] & 0x80) >> 7)
	p.pix[5] = s[8] & 0x7f
	p.pix[6] = (s[9] & 0xfe) >> 1
	p.pix[7] = ((s[9] & 0x01) << 6) + ((s[10] & 0xfc) >> 2)
	p.pix[8] = ((s[10] & 0x03) << 5) + ((s[11] & 0xf8) >> 3)
	p.pix[9] = ((s[11] & 0x07) << 4) + ((s[12] & 0xf0) >> 4)
	p.pix[10] = ((s[12] & 0x0f) << 3) + ((s[13] & 0xe0) >> 5)
	p.pix[11] = ((s[13] & 0x1f) << 2) + ((s[14] & 0xc0) >> 6)
	p.pix[12] = ((s[14] & 0x3f) << 1) + ((s[15] & 0x80) >> 7)
	p.pix[13] = s[15] & 0x7f

	return p
}

//TODO(sjon): spec claims I should handle NULLs for ASCII
func readByte(r io.Reader) byte {
	var bt byte
	binary.Read(r, b, &bt)
	return bt
}

func readUint16(r io.Reader) uint16 {
	var short uint16
	binary.Read(r, b, &short)
	return short
}

func readUint32(r io.Reader) uint32 {
	var long uint32
	binary.Read(r, b, &long)
	return long
}

//Used for fixpoint of 32 bit numerator and denominator
func readUint64(r io.Reader) uint64 {
	var longlong uint64
	binary.Read(r, b, &longlong)
	return longlong
}

func readInt32(r io.Reader) int32 {
	var long int32
	binary.Read(r, b, &long)
	return long
}

//Used for fixpoint of 32 bit numerator and denominator
func readInt64(r io.Reader) int64 {
	var longlong int64
	binary.Read(r, b, &longlong)
	return longlong
}
