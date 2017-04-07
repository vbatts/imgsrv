package dbutil

import (
	"github.com/vbatts/imgsrv/types"
	"labix.org/v2/mgo"
)

// Handles are all the register backing Handlers
var Handles map[string]Handler

// Handler is the means of getting "files" from the backing database
type Handler interface {
	Init(config interface{}) error
	Close() error

	Open(filename string) (*mgo.GridFile, error)
	Create(filename string) (*mgo.GridFile, error)
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
