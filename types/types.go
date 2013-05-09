package types

import (
	"mime"
	"path/filepath"
	"strings"
	"time"
  "fmt"
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
	IsImage     bool
	ContentType string "contentType,omitempty"
}

func (f *File) SetIsImage() {
	m_type := mime.TypeByExtension(filepath.Ext(f.Filename))
	f.IsImage = strings.Contains(m_type, "image")
  fmt.Println(f.Filename,f.IsImage)
}
