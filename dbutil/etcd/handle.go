package etcd

import (
	"encoding/json"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/vbatts/imgsrv/dbutil"
	"github.com/vbatts/imgsrv/types"
)

func init() {
	dbutil.Handles["etcd"] = &eHandle{}
}

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
	// This is going to require a wild helper to
	return nil
}

func (e *eHandle) Close() error {
	return nil
}

func (e *eHandle) Open(filename string) (dbutil.File, error) {
	return nil, nil
}
func (e *eHandle) Create(filename string) (dbutil.File, error) {
	return nil, nil
}
func (e *eHandle) Remove(filename string) error {
	return nil
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
