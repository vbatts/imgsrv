package minio

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/minio/minio-go"
	"github.com/vbatts/imgsrv/dbutil"
	"github.com/vbatts/imgsrv/types"
)

type minioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Secure          bool
	Region          string

	Bucket string
}

func init() {
	dbutil.Handles["minio"] = &minioHandle{}
}

type minioHandle struct {
	Config minioConfig
	Client *minio.Client
}

func (mh *minioHandle) Init(config []byte, err error) error {
	if err != nil {
		return errwrap.Wrapf("minio Init: {{er}}", err)
	}

	mh.Config = minioConfig{}
	if err := json.Unmarshal(config, &mh.Config); err != nil {
		return err
	}

	mh.Client, err = minio.NewWithRegion(mh.Config.Endpoint, mh.Config.AccessKeyID, mh.Config.SecretAccessKey, mh.Config.Secure, mh.Config.Region)
	if err != nil {
		return errwrap.Wrapf("minio Client: {{er}}", err)
	}

	if buc, err := mh.Client.BucketExists(mh.Config.Bucket); err != nil || !buc {
		return fmt.Errorf("can not access bucket %q", mh.Config.Bucket)
	}
	return nil
}

// ErrNotImplemented for functions as the stubs are being written
var ErrNotImplemented = fmt.Errorf("function not implemented")

func (mh *minioHandle) Close() error {
	return nil
}

func (mh *minioHandle) Open(filename string) (dbutil.File, error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) Create(filename string) (dbutil.File, error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) Remove(filename string) error {
	return mh.Client.RemoveObject(mh.Config.Bucket, filename)
}

func (mh *minioHandle) HasFileByFilename(filename string) (exists bool, err error) {
	return false, ErrNotImplemented
}
func (mh *minioHandle) FindFilesByKeyword(keyword string) (files []types.File, err error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) FindFilesByMd5(md5 string) (files []types.File, err error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) FindFilesByPatt(filenamePat string) (files []types.File, err error) {
	return nil, ErrNotImplemented
}

func (mh *minioHandle) CountFiles(filename string) (int, error) {
	return -1, ErrNotImplemented
}

func (mh *minioHandle) GetFiles(limit int) (files []types.File, err error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) GetFileByFilename(filename string) (types.File, error) {
	return types.File{}, ErrNotImplemented
}
func (mh *minioHandle) GetExtensions() (kp []types.IdCount, err error) {
	return nil, ErrNotImplemented
}
func (mh *minioHandle) GetKeywords() (kp []types.IdCount, err error) {
	return nil, ErrNotImplemented
}
