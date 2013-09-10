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
	Metadata    Info ",omitempty"
	Md5         string
	ChunkSize   int
	UploadDate  time.Time
	Length      int64
	Filename    string ",omitempty"
	ContentType string "contentType,omitempty"
}

func (f *File) SetContentType() {
	f.ContentType = mime.TypeByExtension(filepath.Ext(f.Filename))
}

func (f *File) IsImage() bool {
	f.SetContentType()
	return strings.HasPrefix(f.ContentType, "image")
}

func (f *File) IsVideo() bool {
	f.SetContentType()
	return strings.HasPrefix(f.ContentType, "video")
}

func (f *File) IsAudio() bool {
	f.SetContentType()
	return strings.HasPrefix(f.ContentType, "audio")
}

/*
Structure used for collecting values from mongo for a tag cloud
*/
type IdCount struct {
	Id    string "_id"
	Value int
	Root  string
}
