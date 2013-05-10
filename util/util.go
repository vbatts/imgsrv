package util

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func FetchFileFromURL(url string) (filename string, err error) {
	var t time.Time

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		//CheckRedirect: redirectPolicyFunc,
		Transport: tr,
	}
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	mtime := resp.Header.Get("last-modified")
	if len(mtime) > 0 {
		t, err = time.Parse(http.TimeFormat, mtime)
		if err != nil {
			return
		}
	} else {
		log.Println("Last-Modified not present. Using current time")
		t = time.Now()
	}
	_, url_filename := filepath.Split(url)

	log.Println(resp)

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filepath.Join(os.TempDir(), url_filename), bytes, 0644)
	if err != nil {
		return
	}
	err = os.Chtimes(filepath.Join(os.TempDir(), url_filename), t, t)

	// lastly, return
	return filepath.Join(os.TempDir(), url_filename), nil
}
