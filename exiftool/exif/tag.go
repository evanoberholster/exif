package exif

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Errors
var (
	ErrEmptyTag = errors.New("Error empty tag")
)

// ExifReader interface
type ExifReader interface {
	io.ReaderAt
	ByteOrder() binary.ByteOrder
}

// TagID is the uint16 representation of an IFD tag
type TagID uint16

// TagMap - is a lookupmap for Tags using their TagID
type TagMap map[TagID]Tag

// Tag - is an Exif Tag
type Tag struct {
	Name    string
	tagType TagType

	// ValueContext
	ifdPath        string
	tagID          TagID
	unitCount      uint32
	valueOffset    uint32
	rawValueOffset []byte
}

// NewTag -
func NewTag(name string, tagType TagType) Tag {
	return Tag{
		Name:    name,
		tagType: tagType,
	}
}

// Set tag
func (tag *Tag) Set(ifdPath string, tagID TagID, unitCount uint32, valueOffset uint32, rawValueOffset []byte) {
	tag.ifdPath = ifdPath
	tag.tagID = tagID
	tag.unitCount = unitCount
	tag.valueOffset = valueOffset
	tag.rawValueOffset = rawValueOffset
}

// effectiveValueType returns the effective type of the unknown-type tag or, if
// not unknown, the actual type.
func (tag *Tag) effectiveValueType() (tagType TagType) {
	//if tag.tagType == TypeUndefined {
	//	tagType = vc.undefinedValueTagType
	//
	//	if tagType == 0 {
	//		panic(fmt.Errorf("undefined-value type not set"))
	//	}
	//} else {
	tagType = tag.tagType
	//}

	return tagType
}

// readRawEncoded returns the encoded bytes for the value that we represent.
func (tag Tag) readRawEncoded(exifReader io.ReaderAt) (rawBytes []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
		}
	}()

	//tagType := tag.effectiveValueType()

	byteLength := uint32(tag.tagType.Size()) * tag.unitCount

	// if isEmbedded
	if byteLength <= 4 {
		return tag.rawValueOffset[:byteLength], nil
	}

	data := make([]byte, byteLength)
	if _, err = exifReader.ReadAt(data, int64(tag.valueOffset)); err != nil {
		panic(err)
	}
	return data, nil
}

// GetString -
// WIP
func (tag Tag) GetString(exifReader io.ReaderAt) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
			value = ""
		}
	}()
	if !tag.tagType.IsValid() {
		return "", ErrEmptyTag
	}
	rawValue, err := tag.readRawEncoded(exifReader)
	if err != nil {
		return "", err
	}
	switch tag.tagType {
	case TypeASCII:
		value, err = parser.ParseASCII(rawValue, tag.unitCount)
	case TypeASCIINoNul:
		value, err = parser.ParseASCIINoNul(rawValue, tag.unitCount)
	case TypeByte:
		if len(rawValue) < int(TypeByteSize*tag.unitCount) {
			panic(ErrNotEnoughData)
		}
		value = string(bytes.Trim(rawValue[:tag.unitCount], "\x00"))
	default:
		panic(ErrUnparseableValue)
	}
	return
}

// GetInt -
// WIP
func (tag Tag) GetInt(exifReader ExifReader) (value int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
		}
	}()
	if !tag.tagType.IsValid() {
		panic(ErrEmptyTag)
	}
	rawValue, err := tag.readRawEncoded(exifReader)
	if err != nil {
		panic(err)
	}

	if len(rawValue) < tag.tagType.Size()*int(tag.unitCount) {
		panic(ErrNotEnoughData)
	}

	switch tag.tagType {
	case TypeShort:
		value = int(exifReader.ByteOrder().Uint16(rawValue[:2]))
	case TypeLong:
		value = int(exifReader.ByteOrder().Uint32(rawValue[:4]))
	case TypeByte:
		value = int([]uint8(rawValue[:1])[0])
	default:
		panic(ErrUnparseableValue)
	}
	return
}

// GetRational -
// WIP
func (tag Tag) GetRational(exifReader ExifReader) (numerator int, denominator int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
		}
	}()
	if !tag.tagType.IsValid() {
		panic(ErrEmptyTag)
	}
	rawValue, err := tag.readRawEncoded(exifReader)
	if err != nil {
		panic(err)
	}

	byteOrder := exifReader.ByteOrder()
	if len(rawValue) < tag.tagType.Size()*int(tag.unitCount) {
		panic(ErrNotEnoughData)
	}

	switch tag.tagType {
	case TypeRational:
		numerator = int(byteOrder.Uint32(rawValue[0:4]))
		denominator = int(byteOrder.Uint32(rawValue[4:8]))
	case TypeSignedRational:
		numerator = int(byteOrder.Uint32(rawValue[0:4]))
		denominator = int(byteOrder.Uint32(rawValue[4:8]))
	default:
		panic(ErrUnparseableValue)
	}
	return
}

// GetRationals -
// WIP
func (tag Tag) GetRationals(exifReader ExifReader) (value []Rational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
		}
	}()
	if !tag.tagType.IsValid() {
		panic(ErrEmptyTag)
	}
	rawValue, err := tag.readRawEncoded(exifReader)
	if err != nil {
		panic(err)
	}

	byteOrder := exifReader.ByteOrder()
	count := int(tag.unitCount)
	if len(rawValue) < tag.tagType.Size()*count {
		panic(ErrNotEnoughData)
	}

	switch tag.tagType {
	case TypeRational:
		value = make([]Rational, count)
		for i := 0; i < count; i++ {
			value[i].Numerator = byteOrder.Uint32(rawValue[i*8:])
			value[i].Denominator = byteOrder.Uint32(rawValue[i*8+4:])
		}
	default:
		panic(ErrUnparseableValue)
	}
	return
}

// GetUint16 -
// WIP
func (tag Tag) GetUint16(exifReader ExifReader) (values []uint16, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = state.(error)
		}
	}()
	if !tag.tagType.IsValid() {
		panic(ErrEmptyTag)
	}
	rawValue, err := tag.readRawEncoded(exifReader)
	if err != nil {
		panic(err)
	}
	if len(rawValue) < tag.tagType.Size()*int(tag.unitCount) {
		panic(ErrNotEnoughData)
	}

	byteOrder := exifReader.ByteOrder()

	switch tag.tagType {
	case TypeShort:
		values = make([]uint16, tag.unitCount)
		for i := 0; i < int(tag.unitCount); i++ {
			values[i] = byteOrder.Uint16(rawValue[i*2:])
		}
	default:
		panic(ErrUnparseableValue)
	}
	return
}
