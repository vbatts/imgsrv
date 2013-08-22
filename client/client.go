package client

import (
	"bufio"
	"io/ioutil"
	"mime"
	"net/http"
  "net/url"
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

func hurr() {
	values := make(url.Values)
	values.Set("email", "anything@email.com")
	values.Set("name", "bob")
	values.Set("count", "1")
	r, err := http.PostForm("http://example.com/form", values)
	if err != nil {
		return
	}
  _ = r
}
