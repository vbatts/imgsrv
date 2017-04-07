package etcd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/vbatts/imgsrv/dbutil"
	"github.com/vbatts/imgsrv/types"
	"golang.org/x/net/context"
)

func init() {
	dbutil.Handles["etcd"] = &eHandle{}
}

var (
	filesPrefix = "/files/"
	objPrefix   = "/obj/"
	refPrefix   = "/refcount/"
)

type dbConfig struct {
	Endpoints []string
}

type eHandle struct {
	config dbConfig
	c      client.Client
	kapi   client.KeysAPI
}

func (e *eHandle) Init(config []byte, err error) error {
	if err != nil {
		return err
	}
	e.config = dbConfig{}
	if err := json.Unmarshal(config, &e.config); err != nil {
		return err
	}

	cfg := client.Config{
		Endpoints:               e.config.Endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	e.c, err = client.New(cfg)
	if err != nil {
		return err
	}
	e.kapi = client.NewKeysAPI(e.c)
	return nil
}

func (e *eHandle) Close() error {
	return nil
}

func (e *eHandle) Open(filename string) (dbutil.File, error) {
	// This is going to require a wild helper to stash read the file contents
	// from the store, perhaps base64 encoded blob at /obj/md5/<blob>, then at
	// /files/<filename> it is a marshalled object with the md5 sum reference
	// plus additional metadata for the file (i.e. types.File and types.Info)
	resp, err := e.kapi.Get(context.Background(), filesPrefix+filename, nil)
	if err != nil {
		return nil, err
	}
	// unmarshal the data
	fi := types.File{}
	if err := json.Unmarshal([]byte(resp.Node.Value), &fi); err != nil {
		return nil, err
	}

	// then get the object blob
	resp, err = e.kapi.Get(context.Background(), objPrefix+fi.Md5, nil)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(resp.Node.Value)
	if err != nil {
		return nil, err
	}

	// then write obj to this file
	fh, err := ioutil.TempFile("", "imgsrv."+filename)
	if err != nil {
		return nil, err
	}
	if _, err := fh.Write(decoded); err != nil {
		return nil, err
	}
	fh.Sync()
	fh.Seek(0, 0)
	return &eFile{h: e, fh: fh, info: fi}, nil
}

func (e *eHandle) Create(filename string) (dbutil.File, error) {
	// This is will have some similarities to Open(), but will have to buffer the
	// file (bytes.Buffer or ioutil.TempFile). This will have to be in a
	// goroutine that does a checksum and pushes it to the backend on .Close() of
	// the returned File. :-\

	fh, err := ioutil.TempFile("", "imgsrv."+filename)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &eFile{h: e, fh: fh, info: types.File{Filename: filename, UploadDate: now, Metadata: types.Info{TimeStamp: now}}}, nil
}

func (e *eHandle) Remove(filename string) error {
	// Perhaps a little tricky, since you can remove the /files/<filename>, but
	// the /obj/md5/<blob> needs a ref counter, so that blob can be ejected when
	// it has no refs.
	resp, err := e.kapi.Get(context.Background(), filesPrefix+filename, nil)
	if err != nil {
		return err
	}
	fi := types.File{}
	if err := json.Unmarshal([]byte(resp.Node.Value), &fi); err != nil {
		return err
	}
	if _, err := e.kapi.Delete(context.Background(), filesPrefix+filename, nil); err != nil {
		return err
	}
	i, err := e.refCountAdd(fi.Md5, -1)
	if err != nil {
		return err
	}
	if i < 1 {
		if _, err := e.kapi.Delete(context.Background(), objPrefix+fi.Md5, nil); err != nil {
			return err
		}
	}
	return nil
}

// intended for ref counting md5 objects. Add a negative number to decrement
func (e *eHandle) refCountAdd(refname string, i int) (int, error) {
	resp, err := e.kapi.Get(context.Background(), refPrefix+refname, nil)
	if err != nil {
		return -1, err
	}
	count, err := strconv.Atoi(resp.Node.Value)
	if err != nil {
		return -1, err
	}
	_, err = e.kapi.Set(context.Background(), refPrefix+refname, fmt.Sprintf("%d", count+i), nil)
	return count + i, err
}

func (e *eHandle) HasFileByFilename(filename string) (exists bool, err error) {
	return false, nil
}
func (e *eHandle) FindFilesByKeyword(keyword string) (files []types.File, err error) {
	return nil, nil
}
func (e *eHandle) FindFilesByMd5(md5 string) (files []types.File, err error) {
	return nil, nil
}
func (e *eHandle) FindFilesByPatt(filenamePat string) (files []types.File, err error) {
	return nil, nil
}

func (e *eHandle) CountFiles(filename string) (int, error) {
	return 0, nil
}
func (e *eHandle) GetFileByFilename(filename string) (types.File, error) {
	return types.File{}, nil
}
func (e *eHandle) GetFiles(limit int) (files []types.File, err error) {
	return nil, nil
}
func (e *eHandle) GetExtensions() (kp []types.IdCount, err error) {
	return nil, nil
}
func (e *eHandle) GetKeywords() (kp []types.IdCount, err error) {
	return nil, nil
}
