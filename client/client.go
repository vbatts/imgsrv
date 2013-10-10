package client

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

var ErrorNotOK = errors.New("HTTP Response was not 200 OK")

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
	_ = writer.WriteField("returnUrl", "true")
	contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return nil, err
	}

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
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return string(bytes), ErrorNotOK
	}

	return string(bytes), nil
}
