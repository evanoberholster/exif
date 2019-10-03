// Package mknote provides makernote parsers that can be used with goexif/exif.
package mknote

import (
	"bytes"
	"fmt"

	"github.com/evanoberholster/exif/exif"
	"github.com/evanoberholster/exif/tiff"
)

var (
	// Canon is an exif.Parser for canon makernote data.
	Canon = &canon{}
	// NikonV3 is an exif.Parser for nikon makernote data.
	NikonV3 = &nikonV3{}
	// All is a list of all available makernote parsers
	All = []exif.Parser{Canon, NikonV3}
)

type canon struct{}

// Parse decodes all Canon makernote data found in x and adds it to x.
func (_ *canon) Parse(x *exif.Exif) error {
	m, err := x.Get(exif.MakerNote)
	if err != nil {
		return nil
	}

	mk, err := x.Get(exif.Make)
	if err != nil {
		return nil
	}

	if val, err := mk.StringVal(); err != nil || val != "Canon" {
		return nil
	}

	// Canon notes are a single IFD directory with no header.
	// Reader offsets need to be w.r.t. the original tiff structure.
	cReader := bytes.NewReader(append(make([]byte, m.ValOffset), m.Val...))
	cReader.Seek(int64(m.ValOffset), 0)

	mkNotesDir, _, err := tiff.DecodeDir(cReader, x.Tiff.Order)
	if err != nil {
		return err
	}
	x.LoadTags(mkNotesDir, makerNoteCanonFields, false)
	if err := loadSubDir(x, cReader, CanonCameraSettings, makerNoteNikon3PreviewFields); err != nil {
	}
	return nil
}

type nikonV3 struct{}

// Parse decodes all Nikon makernote data found in x and adds it to x.
func (_ *nikonV3) Parse(x *exif.Exif) error {
	m, err := x.Get(exif.MakerNote)
	if err != nil {
		return nil
	}
	if len(m.Val) < 6 {
		return nil
	}
	if bytes.Compare(m.Val[:6], []byte("Nikon\000")) != 0 {
		return nil
	}

	// Nikon v3 maker note is a self-contained IFD (offsets are relative
	// to the start of the maker note)
	nReader := bytes.NewReader(m.Val[10:])
	mkNotes, err := tiff.Decode(nReader)
	if err != nil {
		return err
	}
	makerNoteOffset := m.ValOffset + 10
	x.LoadTags(mkNotes.Dirs[0], makerNoteNikon3Fields, false)

	//if err := loadSubList(x, nReader, NikonPreviewPtr, makerNoteNikon3PreviewFields); err != nil {
	//}
	previewTag, err := x.Get(NikonPreviewImageStart)
	if err == nil {
		offset, _ := previewTag.Int64(0)
		previewTag.SetInt(0, offset+int64(makerNoteOffset))
		x.Update(NikonPreviewImageStart, previewTag)
	}
	fmt.Println("OFFSET.......", m.ValOffset+10)
	fmt.Println(err)

	return nil
}

func loadSubDir(x *exif.Exif, r *bytes.Reader, ptr exif.FieldName, fieldMap map[uint16]exif.FieldName) error {
	tag, err := x.Get(ptr)
	if err != nil {
		return nil
	}
	fmt.Println(tag)
	offset, err := tag.Int64(0)
	if err != nil {
		return nil
	}

	_, err = r.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("exif: seek to sub-IFD %s failed: %v", ptr, err)
	}
	subDir, _, err := tiff.DecodeDir(r, x.Tiff.Order)
	if err != nil {
		return fmt.Errorf("exif: sub-IFD %s decode failed: %v", ptr, err)
	}
	x.LoadTags(subDir, fieldMap, false)
	return nil
}

//func loadSubList(x *exif.Exif, r *bytes.Reader, ptr exif.FieldName, fieldMap map[uint16]exif.FieldName) error {
//	return nil
//}
