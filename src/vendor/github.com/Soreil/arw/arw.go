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
	"reflect"
	"unsafe"
	"bytes"
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
	sshort	  *[]int16
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
	case BYTE,UNDEFINED:
		val = fmt.Sprintf("%x",*f.ascii)
	case ASCII:
		val = fmt.Sprint(string(*f.ascii))
	case SHORT:
		parts := make([]string,len(*f.short))
		for i, short := range *f.short {
			parts[i] = fmt.Sprint(short)
		}
		val = strings.Join(parts,", ")
	case SSHORT:
		parts := make([]string,len(*f.sshort))
		for i, sshort := range *f.sshort {
			parts[i] = fmt.Sprint(sshort)
		}
		val = strings.Join(parts,", ")
	case LONG:
		parts := make([]string,len(*f.long))
		for i, long := range *f.long {
			parts[i] = fmt.Sprint(long)
		}
		val = strings.Join(parts,", ")
	case SLONG:
		parts := make([]string,len(*f.slong))
		for i, slong := range *f.slong {
			parts[i] = fmt.Sprint(slong)
		}
		val = strings.Join(parts,", ")
	case RATIONAL:
		parts := make([]string,len(*f.rat))
		for i, rat := range *f.rat {
			parts[i] = fmt.Sprint(rat)
		}
		val = strings.Join(parts,", ")
	case SRATIONAL:
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
	PhotometricInterpretation IFDtag = 262 //32803: CFA
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

	XMP IFDtag = 700 //http://www.adobe.com/products/xmp.html Some completely useless XML format

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

	BlackLevel                            IFDtag = 0x7300
	WB_GRBGLevelsAuto                     IFDtag = 0x7302
	WB_GRBGLevels                         IFDtag = 0x7303
	BlackLevel2                           IFDtag = 0x7310
	WB_RGGBLevels                         IFDtag = 0x7313
	WB_RGBLevelsDaylight                  IFDtag = 0x7480
	WB_RGBLevelsCloudy                    IFDtag = 0x7481
	WB_RGBLevelsTungsten                  IFDtag = 0x7482
	WB_RGBLevelsFlash                     IFDtag = 0x7483
	WB_RGBLevels4500K                     IFDtag = 0x7484
	WB_RGBLevelsFluorescent               IFDtag = 0x7486
	MaxApertureAtMaxFocal                 IFDtag = 0x74a0
	MaxApertureAtMinFocal                 IFDtag = 0x74a1
	MaxFocalLength                        IFDtag = 0x74a2
	MinFocalLength                        IFDtag = 0x74a3
	SR2DataIFD                            IFDtag = 0x74c0
	ColorMatrix                           IFDtag = 0x7800
	WB_RGBLevelsDaylight2                  IFDtag = 0x7820
	WB_RGBLevelsCloudy2                    IFDtag = 0x7821
	WB_RGBLevelsTungsten2                  IFDtag = 0x7822
	WB_RGBLevelsFlash2                     IFDtag = 0x7823
	WB_RGBLevels4500K2                     IFDtag = 0x7824
	WB_RGBLevelsShade2                     IFDtag = 0x7825
	WB_RGBLevelsFluorescent2               IFDtag = 0x7826
	WB_RGBLevelsFluorescentP1             IFDtag = 0x7827
	WB_RGBLevelsFluorescentP2             IFDtag = 0x7828
	WB_RGBLevelsFluorescentM1             IFDtag = 0x7829
	WB_RGBLevels8500K                     IFDtag = 0x782a
	WB_RGBLevels6000K                     IFDtag = 0x782b
	WB_RGBLevels3200K                     IFDtag = 0x782c
	WB_RGBLevels2500K                     IFDtag = 0x782d
	WhiteLevel                            IFDtag = 0x787f
	VignettingCorrParams                  IFDtag = 0x797d
	ChromaticAberrationCorrParams         IFDtag = 0x7980
	DistortionCorrParams                  IFDtag = 0x7982


	ExifTag             IFDtag = 34665
	GPSTag              IFDtag = 34853
	InteroperabilityTag IFDtag = 40965
	PrintImageMatching  IFDtag = 50341
	DefaultCropOrigin 	IFDtag = 50719
	DefaultCropSize 	IFDtag = 50720
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

//go:generate stringer -type=sonyRawFile
type sonyRawFile uint16

const (
	raw14 sonyRawFile = iota
	raw12
	craw
	crawLossless
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
	SRATIONAL
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
	case RATIONAL, SRATIONAL:
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
			case UNDEFINED,ASCII,BYTE:
				values := make([]byte, interop.Count)
				for i := range values {
					values[i] = byte(((interop.Offset << uint32(8*i)) & 0xff000000) >> 24)
				}
				meta.FIAvals[n].ascii = &values
			case SHORT:
				values := make([]uint16, interop.Count)
				for i := range values {
					values[i] = uint16(((interop.Offset << uint32(16*i)) & 0xffff0000) >> 16)
				}
				meta.FIAvals[n].short = &values
			case SSHORT:
				values := make([]int16, interop.Count)
				for i := range values {
					values[i] = int16(((interop.Offset << uint32(16*i)) & 0xffff0000) >> 16)
				}
				meta.FIAvals[n].sshort = &values
			case LONG:
				values := []uint32{interop.Count}
				meta.FIAvals[n].long = &values
			case SLONG:
				values := []int32{int32(interop.Count)}
				meta.FIAvals[n].slong = &values
			}
		} else {
			r.Seek(int64(interop.Offset), 0)
			switch interop.Type {
			case UNDEFINED,ASCII,BYTE:
				values := make([]byte, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].ascii = &values
			case SHORT:
				values := make([]uint16, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].short = &values
			case SSHORT:
				values := make([]int16, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].sshort = &values
			case LONG:
				values := make([]uint32, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].long = &values
			case SLONG:
				values := make([]int32, interop.Count)
				binary.Read(r, b, &values)
				meta.FIAvals[n].slong = &values
			case RATIONAL:
				values := make([]uint32, interop.Count*2)
				binary.Read(r, b, &values)
				floats := make([]float32,interop.Count)
				for i := range floats {
					floats[i] = float32(values[i*2])/float32(values[(i*2)+1])
				}
				meta.FIAvals[n].rat = &floats
			case SRATIONAL:
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

var pad = []uint32{0xae567acf, 0x3758b80d, 0x7c2906a5, 0x1a30e50c, 0xa4fff8d4, 0x5ad0ba02, 0xb0adfde3, 0x80c0bf1c, 0x28a40a6e, 0xb5210a3c, 0x3013ee1b, 0x6ac26b41, 0x306ec9eb, 0xbfc7c3fa, 0x01fa4ee0, 0xaa0b5077, 0x63280f17, 0x2b98271b, 0xc4a483ee, 0x0327efd8, 0x4f1919f3, 0x507e9187, 0x167b353b, 0xa7b2fcbe, 0xb2c45890, 0xef99db72, 0x497fdb56, 0x91564e98, 0xf777078d, 0xfd9e2bd5, 0x7c11b8b7, 0xd890cb9a, 0x16cd7e75, 0x4b1cc09f, 0xd4b88d85, 0x2719170a, 0x85ebe6e1, 0xd80aae2b, 0xa2a6d6c8, 0xfe277243, 0x4e9a6052, 0x4d5ab8d1, 0xd9796c35, 0x66fb9425, 0x2fc719ce, 0x574259e8, 0xed7debf6, 0x62729b9b, 0x8475e571, 0x6b6084e7, 0xd2101c0e, 0x12243ef8, 0xaccaf2ff, 0xf388743f, 0xfdb4dde3, 0xc259958e, 0xa3fc5e38, 0x63a2c363, 0xbd9006b7, 0x43f7adda, 0x3dd8b01e, 0x41aadc72, 0x01916c53, 0x04bae250, 0x7892b89b, 0x8b207c44, 0xf206a891, 0x1e353d29, 0x14292114, 0x2b2b82da, 0xcd5f120b, 0x6a3c7ee7, 0xb2ed663e, 0x822ef87b, 0xff64e96a, 0xd0250c39, 0x9a121fa9, 0xa516e885, 0xcbecec87, 0xea66c879, 0xa3fce75d, 0x9fe040f8, 0xd12016b4, 0xeb0c1103, 0xe5b8e3d3, 0xe8d8a3f6, 0x6930ebcf, 0x06a865eb, 0x18111138, 0xdde18c3b, 0xe342f4ef, 0xb793d2a1, 0xf7a7caaf, 0xd4e4bc34, 0x29ca7d80, 0xc6eedc2a, 0xbcdb6e5f, 0x2514c03c, 0x2a2326be, 0xc7f5392c, 0x2cf191c2, 0xc4c3f321, 0x0ca46ff9, 0x066c941b, 0x40aafc77, 0x855fcf74, 0x981c261d, 0x0667b6de, 0xb16db5d5, 0x0771f254, 0x53e22691, 0x022c8814, 0xc41f2789, 0x0abaf480, 0x2ffb0330, 0x112cf928, 0xd7c94972, 0x362c1b50, 0xf0659484, 0x4f00c4f1, 0x4f58bbed, 0xf258be43, 0x7f7b5ed2, 0x7ab1f464, 0x6046ca7f, 0x11d3954e, 0x3e7a285b, 0x00000000}

func DecryptSR2(r io.ReaderAt, offset uint32, length uint32) []byte{
	buf := make([]byte,length)
	r.ReadAt(buf,int64(offset))
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	header.Len /= 4
	header.Cap /= 4
	data := *(*[]uint32)(unsafe.Pointer(&header))

	p := 128
	for i := range data {
		pad[(p-1) & 127] = pad[p & 127] ^ pad[(p+64) & 127]
		data[i] ^= pad[(p-1) & 127]
		p++
	}

	return buf
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

type crawPixelBlock struct {
	max    uint16
	min    uint16
	maxidx uint8
	minidx uint8
	pix    [14]uint8
}

type rawPixelBlock struct {
	pix [16]uint16
}

const pixelBlockSize = 16

type pixel uint16

func (p crawPixelBlock) String() string {
	return fmt.Sprintf("%011b\n%011b\n%04b\n%04b\n%08b", p.max, p.min, p.maxidx, p.minidx, p.pix)
}

func (p crawPixelBlock) Decompress() [pixelBlockSize]pixel {
	var pix [pixelBlockSize]pixel
	factor := uint8(1 << uint8(math.Ceil(math.Log2(float64(p.max-p.min)/128))))
	var ordinary int
	for i := 0; i < pixelBlockSize; i++ {
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

func readCrawBlock(s []byte) crawPixelBlock {
	var p crawPixelBlock

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

func readRawBlock(s []byte) rawPixelBlock {
	var p rawPixelBlock

	r := bytes.NewReader(s)
	for i := range p.pix {
		binary.Read(r, b, &p.pix[i])
	}
	return p
}