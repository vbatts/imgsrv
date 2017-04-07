package dbutil

import (
	"io"

	"github.com/vbatts/imgsrv/types"
)

// Handles are all the register backing Handlers
var Handles map[string]Handler

// Handler is the means of getting "files" from the backing database
type Handler interface {
	Init(config interface{}) error
	Close() error

	Open(filename string) (File, error)
	Create(filename string) (File, error)
	Remove(filename string) error

	//HasFileByMd5(md5 string) (exists bool, err error)
	//HasFileByKeyword(keyword string) (exists bool, err error)
	HasFileByFilename(filename string) (exists bool, err error)
	FindFilesByKeyword(keyword string) (files []types.File, err error)
	FindFilesByMd5(md5 string) (files []types.File, err error)
	FindFilesByPatt(filenamePat string) (files []types.File, err error)

	CountFiles(filename string) (int, error)

	GetFiles(limit int) (files []types.File, err error)
	GetFileByFilename(filename string) (types.File, error)
	GetExtensions() (kp []types.IdCount, err error)
	GetKeywords() (kp []types.IdCount, err error)
}

type MetaDataer interface {
	GetMeta(result interface{}) (err error)
	SetMeta(metadata interface{})
}

type File interface {
	io.Reader
	io.Writer
	io.Closer
	MetaDataer
}
