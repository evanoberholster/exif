package exiftool

import (
	"github.com/evanoberholster/exiftools/exiftool/exif"
	"github.com/evanoberholster/exiftools/exiftool/tags/ifd"
)

// NewIfdMapping creates an empty IfdMapping from the rootNode
func NewIfdMapping() (ifdMapping *IfdMapping) {
	rootNode := &MappedIfd{
		Path:     make([]string, 0),
		Children: make(map[exif.TagID]*MappedIfd),
	}

	return &IfdMapping{
		rootNode: rootNode,
	}
}

// MappedIfd -
type MappedIfd struct {
	ParentTagID exif.TagID
	Placement   ifd.IfdPath
	Path        []string

	Name     string
	TagID    exif.TagID
	Children map[exif.TagID]*MappedIfd
}
