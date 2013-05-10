package client

import (
	"bufio"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func PutFileFromPath(host, filename string) (path string, err error) {
	ext := filepath.Ext(filename)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	resp, err := http.Post(host, mime.TypeByExtension(ext), bufio.NewReader(file))
	if err != nil {
		return
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return string(bytes), nil
}
