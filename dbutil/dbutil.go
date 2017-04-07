package dbutil

import (
	"io"

	"github.com/vbatts/imgsrv/types"
)

// Handles are all the register backing Handlers
var Handles = map[string]Handler{}

// Handler is the means of getting "files" from the backing database
type Handler interface {
	Init(config []byte, err error) error
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

// File is what is stored and fetched from the backing database
type File interface {
	io.Reader
	io.Writer
	io.Closer
	MetaDataer
}

// MetaDataer allows set/get for optional metadata
type MetaDataer interface {
	/*
		GetMeta unmarshals the optional "metadata" field associated with the file into
		the result parameter. The meaning of keys under that field is user-defined. For
		example:

			result := struct{ INode int }{}
			err = file.GetMeta(&result)
			if err != nil {
				panic(err.String())
			}
			fmt.Printf("inode: %d\n", result.INode)
	*/
	GetMeta(result interface{}) (err error)
	/*
		SetMeta changes the optional "metadata" field associated with the file.  The
		meaning of keys under that field is user-defined. For example:

			file.SetMeta(bson.M{"inode": inode})

		It is a runtime error to call this function when the file is not open for
		writing.

	*/
	SetMeta(metadata interface{})
}
