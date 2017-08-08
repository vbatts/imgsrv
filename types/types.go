package types

import (
	"mime"
	"path/filepath"
	"strings"
	"time"
)

type Info struct {
	Keywords  []string // tags
	Ip        string   // who uploaded it
	Random    int64
	TimeStamp time.Time "timestamp,omitempty"
}

type File struct {
	Metadata   Info ",omitempty"
	Md5        string
	ChunkSize  int
	UploadDate time.Time
	Length     uint64
	Filename   string ",omitempty"
}

// ContentType guesses the mime-type by the file's extension
func (f *File) ContentType() string {
	return mime.TypeByExtension(filepath.Ext(f.Filename))
}

func (f *File) IsImage() bool {
	return strings.HasPrefix(f.ContentType(), "image")
}

func (f *File) IsVideo() bool {
	if strings.HasSuffix(f.Filename, ".webm") {
		return true
	}
	return strings.HasPrefix(f.ContentType(), "video")
}

func (f *File) IsAudio() bool {
	return strings.HasPrefix(f.ContentType(), "audio")
}

// IdCount structure used for collecting values for a tag cloud
type IdCount struct {
	Id    string "_id"
	Value int
	Root  string
}
