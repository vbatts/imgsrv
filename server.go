package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vbatts/go-httplog"
	"github.com/vbatts/imgsrv/assets"
	"github.com/vbatts/imgsrv/config"
	"github.com/vbatts/imgsrv/dbutil"
	"github.com/vbatts/imgsrv/hash"
	"github.com/vbatts/imgsrv/types"
	"github.com/vbatts/imgsrv/util"
	"labix.org/v2/mgo"
)

var (
	defaultPageLimit int = 25
	serverConfig     config.Config

	mongo_session *mgo.Session  // FIXME make this not global
	images_db     *mgo.Database // FIXME make this not global
	du            dbutil.Util
)

/*
Run as the file/image server
*/
func runServer(c *config.Config) {
	serverConfig = *c

	du = dbutil.Util{
		Seed:   serverConfig.MongoHost,
		User:   serverConfig.MongoUsername,
		Pass:   serverConfig.MongoPassword,
		DbName: serverConfig.MongoDbName,
	}
  err := du.Init()
  if err != nil {
    log.Fatal(err)
  }
	defer du.Close() // TODO this ought to catch a signal to cleanup

	http.HandleFunc("/", routeRoot)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		httplog.DefaultFavIcon.ServeHTTP(w, r)
	})
	http.HandleFunc("/assets/", routeAssets)
	http.HandleFunc("/upload", routeUpload)
	http.HandleFunc("/urlie", routeGetFromUrl)
	http.HandleFunc("/all", routeAll)
	http.HandleFunc("/f/", routeFiles)
	http.HandleFunc("/v/", routeViews)
	http.HandleFunc("/k/", routeKeywords)
	http.HandleFunc("/md5/", routeMD5s)
	http.HandleFunc("/ext/", routeExt)
	http.HandleFunc("/ip/", routeIPs)

	addr := fmt.Sprintf("%s:%s", c.Ip, c.Port)
	log.Printf("Serving on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func initMongo() {
	mongo_session, err := mgo.Dial(serverConfig.MongoHost)
	if err != nil {
		log.Panic(err)
	}
	images_db = mongo_session.DB(serverConfig.MongoDbName)
	if len(serverConfig.MongoUsername) > 0 && len(serverConfig.MongoPassword) > 0 {
		err = images_db.Login(serverConfig.MongoUsername, serverConfig.MongoPassword)
		if err != nil {
			log.Panic(err)
		}
	}
	du.Gfs = images_db.GridFS("fs")
}

func serverErr(w http.ResponseWriter, r *http.Request, e error) {
	httplog.LogRequest(r, 503)
	log.Printf("Error: %s", e)
	w.WriteHeader(503)
	//ErrorPage(w, err)
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

func routeViewsGET(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if len(uriChunks) == 2 && len(uriChunks[1]) > 0 {
		file, err := du.GetFileByFilename(uriChunks[1])
		if err != nil {
			serverErr(w, r, err)
			return
		}
		err = ImageViewPage(w, file)
		if err != nil {
			log.Printf("error: %s", err)
		}

	} else {
		// no filename given, show them the full listing
		http.Redirect(w, r, "/all", 302)
	}
	httplog.LogRequest(r, 200)
}

/*
  GET /f/
  GET /f/:name
*/
// Show a page of most recent images, and tags, and uploaders ...
func routeFilesGET(w http.ResponseWriter, r *http.Request) {
	var err error

	uriChunks := chunkURI(r.URL.Path)
	if len(uriChunks) > 2 {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		serverErr(w, r, err)
		return
	}

	filename := strings.ToLower(uriChunks[1])

	// if the Request got here by a delete request, confirm it
	if (len(r.Form["delete"]) > 0 && r.Form["delete"][0] == "true") && (len(r.Form["confirm"]) > 0 && r.Form["confirm"][0] == "true") {
		httplog.LogRequest(r, 200)
		routeFilesDELETE(w, r)
		return
	} else if len(r.Form["delete"]) > 0 && r.Form["delete"][0] == "true" {
		httplog.LogRequest(r, 200)
		err = DeleteFilePage(w, filename)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		return
	}

	if len(uriChunks) == 2 && len(filename) > 0 {
		log.Printf("Searching for [%s] ...", filename)
		c, err := du.CountFiles(filename)
		// preliminary checks, if they've passed an image name
		if err != nil {
			serverErr(w, r, err)
			return
		}
		log.Printf("Results for [%s] = %d", filename, c)
		if c == 0 {
			httplog.LogRequest(r, 404)
			http.NotFound(w, r)
			return
		}

		ext := filepath.Ext(filename)
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		w.Header().Set("Cache-Control", "max-age=315360000")
		w.WriteHeader(http.StatusOK)

		file, err := du.Open(filename)
		if err != nil {
			serverErr(w, r, err)
			return
		}

		io.Copy(w, file) // send the contents of the file in the body
	} else {
		// no filename given, show them the full listing
		http.Redirect(w, r, "/all", 302)
	}
	httplog.LogRequest(r, 200)
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
		httplog.LogRequest(r, 403)
		http.Error(w, "Not Acceptable", 403)
		return
	}

	var filename string
	info := types.Info{
		Ip:        r.RemoteAddr,
		Random:    hash.Rand64(),
		TimeStamp: time.Now(),
	}

	filename = r.FormValue("filename")
	if len(filename) == 0 && len(uriChunks) == 2 && len(uriChunks[1]) != 0 {
		filename = strings.ToLower(uriChunks[1])
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
				for _, word := range strings.Split(v, ",") {
					info.Keywords = append(info.Keywords, strings.Trim(word, " "))
				}
			} else {
				info.Keywords = append(info.Keywords, strings.Trim(v, " "))
			}
		}
	}

	if len(filename) == 0 {
		str := hash.GetSmallHash()
		if len(p_ext) == 0 {
			filename = fmt.Sprintf("%s.jpg", str)
		} else {
			filename = fmt.Sprintf("%s%s", str, p_ext)
		}
	}

	exists, err := du.HasFileByFilename(filename)
	if err == nil && !exists {
		file, err := du.Create(filename)
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
			file, err := du.Open(filename)
			defer file.Close()
			if err != nil {
				serverErr(w, r, err)
				return
			}

			var mInfo types.Info
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

	httplog.LogRequest(r, 200)
}

func routeFilesPUT(w http.ResponseWriter, r *http.Request) {
	// update the file by the name in the path and/or parameter?
	// update/add keywords from the parameters
	// look for an image in the r.Body
	httplog.LogRequest(r, 418)
}

func routeFilesDELETE(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if (len(uriChunks) > 2) || (len(uriChunks) == 2 && len(uriChunks[1]) == 0) {
		httplog.LogRequest(r, 400)
		http.Error(w, "Bad Syntax", 400)
		return
	}

	exists, err := du.HasFileByFilename(uriChunks[1])
	if err != nil {
		serverErr(w, r, err)
		return
	}

	if exists {
		err = du.Remove(uriChunks[1])
		if err != nil {
			serverErr(w, r, err)
			return
		}
		httplog.LogRequest(r, 302)
		http.Redirect(w, r, "/", 302)
	} else {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
	}
	// delete the name in the path and/or parameter?
}

func routeViews(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		routeViewsGET(w, r)
	default:
		httplog.LogRequest(r, 404)
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
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
}

func routeRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	// Show a page of most recent images, and tags, and uploaders ...

	w.Header().Set("Content-Type", "text/html")
	var files []types.File
	files, err := du.GetFiles(defaultPageLimit)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}
	httplog.LogRequest(r, 200)
}

func routeAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Show a page of all the images
	var files []types.File
	files, err := du.GetFiles(-1)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}
	httplog.LogRequest(r, 200)
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
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) == 1 || (len(uriChunks) == 2 && len(uriChunks[1]) == 0) {
		// Path: /k/
		// show a tag cloud!
		kc, err := du.GetKeywords()
		if err != nil {
			serverErr(w, r, err)
			return
		}
		err = ListTagCloudPage(w, kc)
		if err != nil {
			serverErr(w, r, err)
		}
		return
	}

	log.Printf("K: %s (%d)", uriChunks, len(uriChunks))

	if uriChunks[len(uriChunks)-1] == "r" {
		// Path: /k/
		// TODO determine how to show a random image by keyword ...
		log.Println("random isn't built yet")
		httplog.LogRequest(r, 404)
		return
	}

	var (
		files []types.File
		err   error
	)
	if len(uriChunks) == 2 {
		// Path: /k/:name
		log.Println(uriChunks[1])
		files, err = du.FindFilesByKeyword(uriChunks[1])
		if err != nil {
			serverErr(w, r, err)
			return
		}
	}

	log.Printf("collected %d files", len(files))
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}

	httplog.LogRequest(r, 200)
}

func routeMD5s(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if r.Method != "GET" {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) != 2 {
		// Path: /md5/
		kc, err := du.GetKeywords()
		if err != nil {
			serverErr(w, r, err)
			return
		}
		err = ListTagCloudPage(w, kc)
		if err != nil {
			serverErr(w, r, err)
		}
		return
	}

	files, err := du.FindFilesByMd5(uriChunks[1])
	if err != nil {
		serverErr(w, r, err)
		return
	}
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}

	httplog.LogRequest(r, 200)
}

/*
  GET /ext/
  GET /ext/:name
  GET /ext/:name/r

  Show a page of file extensions, and allow paging by ext
  If /ext/name/r then show a random image by keyword name
  Otherwise 404
*/
func routeExt(w http.ResponseWriter, r *http.Request) {
	uriChunks := chunkURI(r.URL.Path)
	if r.Method != "GET" ||
		len(uriChunks) > 3 ||
		(len(uriChunks) == 3 && uriChunks[2] != "r") {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	} else if len(uriChunks) == 1 || (len(uriChunks) == 2 && len(uriChunks[1]) == 0) {
		// Path: /ext/
		// tag cloud of extensions used
		ic, err := du.GetExtensions()
		if err != nil {
			serverErr(w, r, err)
			return
		}
		log.Printf("ext: %#v", ic)
		err = ListTagCloudPage(w, ic)
		if err != nil {
			serverErr(w, r, err)
		}
		return
	}

	ext := strings.ToLower(uriChunks[1])
	ext_pat := fmt.Sprintf("%s$", ext)
	files, err := du.FindFilesByPatt(ext_pat)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	log.Printf("collected %d files, with ext %s", len(files), ext)
	err = ListFilesPage(w, files)
	if err != nil {
		log.Printf("error: %s", err)
	}

	httplog.LogRequest(r, 200)
}

// Show a page of all the uploader's IPs, and the images
func routeIPs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}

	httplog.LogRequest(r, 200)
}

/*
  GET /urlie
  POST /urlie
*/
func routeGetFromUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := UrliePage(w)
		if err != nil {
			log.Printf("error: %s", err)
		}
		httplog.LogRequest(r, 200)
		return
	}

	if r.Method == "POST" {
		var (
			err             error
			stored_filename string
			local_filename  string
			useRandName     bool = false
			info            types.Info
		)

		info = types.Info{
			Ip:        r.RemoteAddr,
			Random:    hash.Rand64(),
			TimeStamp: time.Now(),
		}
		log.Println(info)

		err = r.ParseMultipartForm(1024 * 5)
		if err != nil {
			serverErr(w, r, err)
			return
		}

		log.Printf("%q", r.MultipartForm.Value)
		for k, v := range r.MultipartForm.Value {
			if k == "keywords" {
				info.Keywords = append(info.Keywords, strings.Split(v[0], ",")...)
			} else if k == "url" {
				local_filename, err = util.FetchFileFromURL(v[0])
				if err != nil {
					serverErr(w, r, err)
					return
				} else if len(local_filename) == 0 {
					httplog.LogRequest(r, 404)
					http.NotFound(w, r)
					return
				}
				// Yay, hopefully we got an image!
			} else if k == "rand" {
				useRandName = true
			} else {
				log.Printf("WARN: not sure what to do with param [%s = %s]", k, v)
			}
		}
		exists, err := du.HasFileByFilename(filepath.Base(strings.ToLower(local_filename)))
		if err != nil {
			serverErr(w, r, err)
			return
		}

		if exists || useRandName {
			ext := filepath.Ext(local_filename)
			str := hash.GetSmallHash()
			stored_filename = fmt.Sprintf("%s%s", str, ext)
		} else {
			stored_filename = filepath.Base(local_filename)
		}

		file, err := du.Create(stored_filename)
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
		log.Printf("Wrote [%d] bytes from %s to %s", n, local_filename, stored_filename)

		http.Redirect(w, r, fmt.Sprintf("/v/%s", stored_filename), 302)
	} else {
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	httplog.LogRequest(r, 200) // if we make it this far, then log success
}

/*
  GET /upload
  POST /upload
*/
func routeUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show the upload form
		httplog.LogRequest(r, 200) // if we make it this far, then log success
		err := UploadPage(w)
		if err != nil {
			log.Printf("error: %s", err)
		}
		return
	}

	if r.Method == "POST" {
		info := types.Info{
			Ip:        r.RemoteAddr,
			Random:    hash.Rand64(),
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
		exists, err := du.HasFileByFilename(filename)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		if exists || useRandName {
			ext := filepath.Ext(filename)
			str := hash.GetSmallHash()
			filename = strings.ToLower(fmt.Sprintf("%s%s", str, ext))
		}

		file, err := du.Create(filename)
		defer file.Close()
		if err != nil {
			log.Printf("Failed to create on gfs: %s", err)
			serverErr(w, r, err)
			return
		}
		file.SetMeta(&info)

		multiFile, err := filehdr.Open()
		if err != nil {
			log.Printf("Failed to open from MultipartForm: %s", err)
			return
		}
		n, err := io.Copy(file, multiFile)
		if err != nil {
			log.Printf("Failed copy from MultipartForm to gfs: %s", err)
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
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	httplog.LogRequest(r, 200) // if we make it this far, then log success
}

func routeAssets(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Rel("/assets", r.URL.Path)
	if err != nil {
		serverErr(w, r, err)
		return
	}

	w.Header().Set("Cache-Control", "max-age=315360000, public")
	w.Header().Set("Expires", time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC1123))

	switch path {
	case "bootstrap.css":
		w.Header().Set("Content-Type", "text/css")
		fmt.Fprintf(w, "%s", assets.BootstrapCss())
	case "bootstrap.js":
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprintf(w, "%s", assets.BootstrapJs())
	case "jquery.js":
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprintf(w, "%s", assets.JqueryJs())
	case "jqud.js":
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprintf(w, "%s", assets.TagCloudJs())
	default:
		httplog.LogRequest(r, 404)
		http.NotFound(w, r)
		return
	}
	httplog.LogRequest(r, 200) // if we make it this far, then log success
}
