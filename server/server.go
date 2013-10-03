package server

import (
  "github.com/gorilla/mux"
	"github.com/vbatts/imgsrv/dbutil"
)

type Web struct {
  Router *mux.Router
  Store  dbutil.Util
}


