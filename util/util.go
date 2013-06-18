package util

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

/* kindof a common log type output */
func LogRequest(r *http.Request, statusCode int) {
	var addr string
	var user_agent string

	user_agent = ""
	addr = r.RemoteAddr

	for k, v := range r.Header {
		if k == "User-Agent" {
			user_agent = strings.Join(v, " ")
		}
		if k == "X-Forwarded-For" {
			addr = strings.Join(v, " ")
		}
	}

	fmt.Printf("%s - - [%s] \"%s %s\" \"%s\" %d %d\n",
		addr,
		time.Now(),
		r.Method,
		r.URL.String(),
		user_agent,
		statusCode,
		r.ContentLength)
}
