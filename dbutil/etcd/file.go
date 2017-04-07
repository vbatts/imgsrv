package etcd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/vbatts/imgsrv/hash"
	"github.com/vbatts/imgsrv/types"
)

// eFile is for wrapping files and satisfy the dbutil.File interface
type eFile struct {
	h          *eHandle
	fh         *os.File
	info       types.File
	hasWritten bool
}

func (f *eFile) Read(p []byte) (n int, err error) {
	if f.fh != nil {
		return f.fh.Read(p)
	}
	return -1, errors.New("no file to read from")
}

func (f *eFile) Write(p []byte) (n int, err error) {
	if f.fh != nil {
		f.hasWritten = true
		return f.fh.Write(p)
	}
	return -1, errors.New("no file to write to")
}

func (f *eFile) Close() error {
	if f.fh != nil {
		if f.hasWritten {
			f.fh.Sync()
			f.fh.Seek(0, 0)
			buf, err := ioutil.ReadAll(f.fh)
			if err != nil {
				return err
			}
			f.info.Md5 = fmt.Sprintf("%x", hash.GetMd5FromBytes(buf))
			_, err = f.h.kapi.Set(context.Background(), objPrefix+f.info.Md5, base64.StdEncoding.EncodeToString(buf), nil)
			if err != nil {
				return err
			}

			buf, err = json.Marshal(f.info)
			if err != nil {
				return err
			}
			_, err = f.h.kapi.Set(context.Background(), filesPrefix+f.info.Filename, string(buf), nil)
			if err != nil {
				return err
			}
			if _, err := f.h.refCountAdd(f.info.Md5, 1); err != nil {
				return err
			}
		}
		if err := f.fh.Close(); err != nil {
			os.Remove(f.fh.Name())
			return err
		}
		return os.Remove(f.fh.Name())
	}
	return nil
}

func (f *eFile) GetMeta(result interface{}) (err error) {
	result = &f.info.Metadata
	return nil
}
func (f *eFile) SetMeta(metadata interface{}) {
	f.info.Metadata = *metadata.(*types.Info)
	f.hasWritten = true
}
