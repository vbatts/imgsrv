package client

import (
	"io/ioutil"
	"net/http"
  "net/url"
  "log"
	"os"
  "mime/multipart"
  "bytes"
  "path"
  "io"
)

func NewfileUploadRequest(uri, file_path string, params map[string]string) (*http.Request, error) {
	file, err := os.Open(file_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("filename", path.Base(file_path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
  contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return nil, err
	}

  log.Println(uri)
  req, err := http.NewRequest("POST", uri, body)
  req.Header.Add("Content-Type", contentType)
  return req, err
}

func PutFileFromPath(uri, file_path string, params map[string]string) (path string, err error) {
  request, err := NewfileUploadRequest(uri, file_path, params)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
  resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
  log.Printf("%#v",resp)
  defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
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
