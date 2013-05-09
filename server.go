package main

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultPageLimit int = 25

/* Run as the image server */
func runServer(ip, port string) {
	var addr = fmt.Sprintf("%s:%s", ip, port)

	initMongo()
	defer mongo_session.Close()

	http.HandleFunc("/", routeRoot)
	http.HandleFunc("/assets/", routeAssets)
	http.HandleFunc("/upload", routeUpload)
	http.HandleFunc("/urlie", routeGetFromUrl)
	http.HandleFunc("/all", routeAll)
	http.HandleFunc("/f/", routeFiles)
	http.HandleFunc("/v/", routeViews)
	http.HandleFunc("/k/", routeKeywords)
	http.HandleFunc("/ip/", routeIPs)
	http.HandleFunc("/ext/", routeExt)
	http.HandleFunc("/md5/", routeMD5s)

	log.Printf("Serving on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func initMongo() {
	mongo_session, err := mgo.Dial(MongoHost)
	if err != nil {
		log.Panic(err)
	}
	images_db = mongo_session.DB(MongoDB)
	if len(MongoUsername) > 0 && len(MongoPassword) > 0 {
		err = images_db.Login(MongoUsername, MongoPassword)
		if err != nil {
			log.Panic(err)
		}
	}
	gfs = images_db.GridFS("fs")
}

func serverErr(w http.ResponseWriter, r *http.Request, e error) {
	log.Printf("Error: %s", e)
	LogRequest(r, 503)
	fmt.Fprintf(w, "Error: %s", e)
	http.Error(w, "Service Unavailable", 503)
	return
}

/* return a <a href/> for a given filename 
   and root is the relavtive base of the explicit link.
*/
func linkToFile(root string, filename string) (html string) {
	return fmt.Sprintf("<a href='%s/f/%s'>%s</a>",
		root,
		filename,
		filename)
}

/* return the sections of the URI Path.
   This will disregard the leading '/'
*/
func chunkURI(uri string) (chunks []string) {
	var str string
	if uri[0] == '/' {
		str = uri[1:]
	} else {
		str = uri
	}
	return strings.Split(str, "/")
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
		r.URL.Path,
		user_agent,
		statusCode,
		r.ContentLength)
}

func routeViewsGET(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if len(uriChunks) == 2 && len(uriChunks[1]) > 0 {
		var file File
		err := gfs.Find(bson.M{"filename": uriChunks[1]}).One(&file)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		file.SetIsImage()
		err = ImageViewPage(w, file)
		if err != nil {
			log.Printf("error: %s", err)
		}

	} else {
		// no filename given, show them the full listing
		http.Redirect(w, r, "/all", 302)
	}
	LogRequest(r, 200)
}

/*
  GET /f/
  GET /f/:name
*/
// Show a page of most recent images, and tags, and uploaders ...
func routeFilesGET(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	if len(uriChunks) == 2 && len(uriChunks[1]) > 0 {
		log.Printf("Searching for [%s] ...", uriChunks[1])
		query := gfs.Find(bson.M{"filename": uriChunks[1]})

		c, err := query.Count()
		// preliminary checks, if they've passed an image name
		if err != nil {
			serverErr(w, r, err)
			return
		}
		log.Printf("Results for [%s] = %d", uriChunks[1], c)
		if c == 0 {
			LogRequest(r, 404)
			http.NotFound(w, r)
			return
		}

		ext := filepath.Ext(uriChunks[1])
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		w.Header().Set("Cache-Control", "max-age=315360000")
		w.WriteHeader(http.StatusOK)

		file, err := gfs.Open(uriChunks[1])
		if err != nil {
			serverErr(w, r, err)
			return
		}

		io.Copy(w, file) // send the contents of the file in the body

	} else {
		// no filename given, show them the full listing
		http.Redirect(w, r, "/all", 302)
	}
	LogRequest(r, 200)
}

/*
  POST /f/[:name][?k=v&k=v]
*/
// Create the file by the name in the path and/or parameter?
// add keywords from the parameters
// look for an image in the r.Body
func routeFilesPOST(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 &&
		((len(uriChunks) == 2 && len(uriChunks[1]) == 0) &&
			len(r.URL.RawQuery) == 0) {
		LogRequest(r, 403)
		http.Error(w, "Not Acceptable", 403)
		return
	}

	var filename string
	info := Info{
		Ip:        r.RemoteAddr,
		Random:    Rand64(),
		TimeStamp: time.Now(),
	}

	filename = r.FormValue("filename")
	if len(filename) == 0 && len(uriChunks) == 2 && len(uriChunks[1]) != 0 {
		filename = uriChunks[1]
	}
	log.Printf("%s\n", filename)

	var p_ext string
	p_ext = r.FormValue("ext")
	if len(filename) > 0 && len(p_ext) == 0 {
		p_ext = filepath.Ext(filename)
	} else if len(p_ext) > 0 && strings.HasPrefix(p_ext, ".") {
		p_ext = fmt.Sprintf(".%s", p_ext)
	}

	for _, word := range []string{
		"k", "key", "keyword",
		"keys", "keywords",
	} {
		v := r.FormValue(word)
		if len(v) > 0 {
			if strings.Contains(v, ",") {
				info.Keywords = append(info.Keywords, strings.Split(v, ",")...)
			} else {
				info.Keywords = append(info.Keywords, v)
			}
		}
	}

	if len(filename) == 0 {
		str := GetSmallHash()
		if len(p_ext) == 0 {
			filename = fmt.Sprintf("%s.jpg", str)
		} else {
			filename = fmt.Sprintf("%s%s", str, p_ext)
		}
	}

	exists, err := HasFileByFilename(filename)
	if err == nil && !exists {
		file, err := gfs.Create(filename)
		defer file.Close()
		if err != nil {
			serverErr(w, r, err)
			return
		}

		file.SetMeta(&info)

		// copy the request body into the gfs file
		n, err := io.Copy(file, r.Body)
		if err != nil {
			serverErr(w, r, err)
			return
		}

		if n != r.ContentLength {
			log.Printf("WARNING: [%s] content-length (%d), content written (%d)",
				filename,
				r.ContentLength,
				n)
		}
	} else if exists {
		if r.Method == "PUT" {
			// TODO nothing will get here presently. Workflow needs more review
			file, err := gfs.Open(filename)
			defer file.Close()
			if err != nil {
				serverErr(w, r, err)
				return
			}

			var mInfo Info
			err = file.GetMeta(&mInfo)
			if err != nil {
				log.Printf("ERROR: failed to get metadata for %s. %s\n", filename, err)
			}
			mInfo.Keywords = append(mInfo.Keywords, info.Keywords...)
			file.SetMeta(&mInfo)
		} else {
			log.Printf("[%s] already exists", filename)
		}
	} else {
		serverErr(w, r, err)
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		io.WriteString(w,
			fmt.Sprintf("<a href=\"/f/%s\">/f/%s</a>\n", filename, filename))
	} else {
		io.WriteString(w, fmt.Sprintf("/f/%s\n", filename))
	}

	LogRequest(r, 200)
}

func routeFilesPUT(w http.ResponseWriter, r *http.Request) {
	// update the file by the name in the path and/or parameter?
	// update/add keywords from the parameters
	// look for an image in the r.Body
	LogRequest(r, 200)
}

func routeFilesDELETE(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) == 2 && len(uriChunks[1]) == 0 {
	}
	exists, err := HasFileByFilename(uriChunks[1])
	if err != nil {
		serverErr(w, r, err)
		return
	}
	if exists {
		err = gfs.Remove(uriChunks[1])
		if err != nil {
			serverErr(w, r, err)
			return
		}
		LogRequest(r, 200)
	} else {
		LogRequest(r, 404)
		http.NotFound(w, r)
	}
	// delete the name in the path and/or parameter?
}

func routeViews(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		routeViewsGET(w, r)
	default:
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
}

func routeFiles(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		routeFilesGET(w, r)
	case r.Method == "PUT":
		routeFilesPUT(w, r)
	case r.Method == "POST":
		routeFilesPOST(w, r)
	case r.Method == "DELETE":
		routeFilesDELETE(w, r)
	default:
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
}

func routeRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	// Show a page of most recent images, and tags, and uploaders ...

	w.Header().Set("Content-Type", "text/html")
	//iter := gfs.Find(bson.M{"uploadDate": bson.M{"$gt": time.Now().Add(-time.Hour)}}).Limit(defaultPageLimit).Iter()
	var files []File
	err := gfs.Find(nil).Sort("-metadata.timestamp").Limit(defaultPageLimit).All(&files)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}
	LogRequest(r, 200)
}

func routeAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Show a page of all the images
	var files []File
	err := gfs.Find(nil).All(&files)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}
	LogRequest(r, 200)
}

/*
  GET /k/
  GET /k/:name
  GET /k/:name/r

  Show a page of all the keyword tags, and then the images
  If /k/:name/r then show a random image by keyword name
  Otherwise 404
*/
func routeKeywords(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if r.Method != "GET" ||
		len(uriChunks) > 3 ||
		(len(uriChunks) == 3 && uriChunks[2] != "r") {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) == 1 || (len(uriChunks) == 2 && len(uriChunks[1]) == 0) {
		routeRoot(w, r)
		return
	}

	log.Printf("K: %s (%d)", uriChunks, len(uriChunks))

	var iter *mgo.Iter
	if uriChunks[len(uriChunks)-1] == "r" {
		// TODO determine how to show a random image by keyword ...
		log.Println("random isn't built yet")
		LogRequest(r, 404)
		return
	} else if len(uriChunks) == 2 {
		log.Println(uriChunks[1])
		iter = gfs.Find(bson.M{"metadata.keywords": uriChunks[1]}).Sort("-metadata.timestamp").Limit(defaultPageLimit).Iter()
	}

	var files []File
	err := iter.All(&files)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	log.Println(len(files))
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}

	LogRequest(r, 200)
}

func routeMD5s(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if r.Method != "GET" {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) != 2 {
		// they didn't give an MD5, re-route
		routeRoot(w, r)
		return
	}

	var files []File
	err := gfs.Find(bson.M{"md5": uriChunks[1]}).Sort("-metadata.timestamp").Limit(defaultPageLimit).All(&files)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}

	LogRequest(r, 200)
}

// Show a page of file extensions, and allow paging by ext
func routeExt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	LogRequest(r, 200)
}

// Show a page of all the uploader's IPs, and the images
func routeIPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	LogRequest(r, 200)
}

/*
  GET /urlie
  POST /urlie
*/
func routeGetFromUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		info := Info{
			Ip:        r.RemoteAddr,
			Random:    Rand64(),
			TimeStamp: time.Now(),
		}
		log.Println(info)

		err := r.ParseMultipartForm(1024 * 5)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		log.Printf("%q", r.MultipartForm.Value)
		var local_filename string
		for k, v := range r.MultipartForm.Value {
			if k == "keywords" {
				info.Keywords = append(info.Keywords, strings.Split(v[0], ",")...)
			} else if k == "url" {
				local_filename, err = FetchFileFromURL(v[0])
				if err != nil {
					serverErr(w, r, err)
					return
				} else if len(local_filename) == 0 {
					LogRequest(r, 404)
					http.NotFound(w, r)
					return
				}
				// Yay, hopefully we got an image!
			} else {
				log.Printf("WARN: not sure what to do with param [%s = %s]", k, v)
			}
		}
		exists, err := HasFileByFilename(local_filename)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		if !exists {
			file, err := gfs.Create(filepath.Base(local_filename))
			defer file.Close()
			if err != nil {
				serverErr(w, r, err)
				return
			}

			local_fh, err := os.Open(local_filename)
			defer local_fh.Close()
			if err != nil {
				serverErr(w, r, err)
				return
			}

			file.SetMeta(&info)

			// copy the request body into the gfs file
			n, err := io.Copy(file, local_fh)
			if err != nil {
				serverErr(w, r, err)
				return
			}
			log.Printf("Wrote [%d] bytes from %s", n, local_filename)

			http.Redirect(w, r, fmt.Sprintf("/v/%s", filepath.Base(local_filename)), 302)
		} else {
			serverErr(w, r, err)
			return
		}
	} else if r.Method == "GET" {
		err := UrliePage(w)
		if err != nil {
			log.Printf("error: %s", err)
		}
	} else {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	LogRequest(r, 200)
}

/*
  GET /upload
  POST /upload
*/
func routeUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show the upload form
		LogRequest(r, 200) // if we make it this far, then log success
		err := UploadPage(w)
		if err != nil {
			log.Printf("error: %s", err)
		}
		return
	}

	if r.Method == "POST" {
		info := Info{
			Ip:        r.RemoteAddr,
			Random:    Rand64(),
			TimeStamp: time.Now(),
		}

		// handle the form posting to this route
		err := r.ParseMultipartForm(1024 * 5)
		if err != nil {
			serverErr(w, r, err)
			return
		}
    useRandName := false
		log.Printf("%q", r.MultipartForm.Value)
		for k, v := range r.MultipartForm.Value {
			if k == "keywords" {
				info.Keywords = append(info.Keywords, strings.Split(v[0], ",")...)
			} else if k == "rand" {
        useRandName = true
			} else {
				log.Printf("WARN: not sure what to do with param [%s = %s]", k, v)
			}
		}

		filehdr := r.MultipartForm.File["filename"][0]
		filename := filehdr.Filename
		exists, err := HasFileByFilename(filename)
		if err != nil {
			serverErr(w, r, err)
			return
		}
    if exists || useRandName {
      ext := filepath.Ext(filename)
			str := GetSmallHash()
      filename = fmt.Sprintf("%s%s", str, ext)
		}

		file, err := gfs.Create(filename)
		defer file.Close()
		if err != nil {
			serverErr(w, r, err)
			return
		}
		file.SetMeta(&info)

		multiFile, err := filehdr.Open()
		if err != nil {
			log.Println(err)
			return
		}
		n, err := io.Copy(file, multiFile)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		if n != r.ContentLength {
			log.Printf("WARNING: [%s] content-length (%d), content written (%d)",
				filename,
				r.ContentLength,
				n)
		}

		http.Redirect(w, r, fmt.Sprintf("/v/%s", filename), 302)
	} else {
		LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	LogRequest(r, 200) // if we make it this far, then log success
}

func routeAssets(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Rel("/assets", r.URL.Path)
	if err != nil {
		serverErr(w, r, err)
		return
	}

	switch path {
	case "bootstrap.css":
		fmt.Fprint(w, bootstrapCSS)
		w.Header().Set("Content-Type", "text/css")
	}
}
