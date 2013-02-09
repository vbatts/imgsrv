package main

import (
  "fmt"
  "io"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "log"
  "math/rand"
  "mime"
  "net/http"
  "path/filepath"
  "strings"
  "time"
)

func serverErr(w http.ResponseWriter, r *http.Request, e error) {
      log.Printf("Error: %s", e)
      LogRequest(r,503)
      fmt.Fprintf(w,"Error: %s", e)
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
  if (uri[0] == '/') {
    str = uri[1:]
  } else {
    str = uri
  }
  return strings.Split(str, "/")
}

/* given an url.URL.RawQuery, get a dictionary in return */
func parseRawQuery(qry string) (params map[string]string) {
  qryChunks := strings.Split(qry, "&")
  params = make(map[string]string, len(qryChunks))
  for _, chunk := range qryChunks {
    p := strings.SplitN(chunk, "=", 2)
    if (len(p) == 2) {
      params[p[0]] = p[1]
    }
  }
  return params
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
      addr = strings.Join(v," ")
    }
  }

  fmt.Printf("%s - - [%s] \"%s %s\" \"%s\" %d %d\n",
    addr,
    time.Now(),
    r.Method,
    r.URL.Path,
    user_agent,
    statusCode,
    r.ContentLength )
}

/*
  GET /f/
  GET /f/:name
*/
// Show a page of most recent images, and tags, and uploaders ...
func routeFilesGET(w http.ResponseWriter, r *http.Request) {
  uriChunks := chunkURI(r.URL.Path)
  if ( len(uriChunks) > 2 ) {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  if (len(uriChunks) == 2 && len(uriChunks[1]) > 0) {
    log.Printf("Searching for [%s] ...", uriChunks[1])
    query := gfs.Find(bson.M{"filename": uriChunks[1] })

    c, err := query.Count()
    // preliminary checks, if they've passed an image name
    if (err != nil) {
      serverErr(w,r,err)
      return
    }
    log.Printf("Results for [%s] = %d", uriChunks[1], c)
    if (c == 0) {
      LogRequest(r,404)
      http.NotFound(w,r)
      return
    }

    ext := filepath.Ext(uriChunks[1])
    w.Header().Set("Content-Type", mime.TypeByExtension(ext))
    w.Header().Set("Cache-Control", "max-age=315360000")
    w.WriteHeader(http.StatusOK)

    file, err := gfs.Open(uriChunks[1])
    if (err != nil) {
      serverErr(w,r,err)
      return
    }

    io.Copy(w,file) // send the contents of the file in the body

  } else {
    // TODO show a list of recent uploads? ...
  }
  LogRequest(r,200)
}

/*
  POST /f/[:name][?k=v&k=v]
*/
// Create the file by the name in the path and/or parameter?
// add keywords from the parameters
// look for an image in the r.Body
func routeFilesPOST(w http.ResponseWriter, r *http.Request) {
  uriChunks := chunkURI(r.URL.Path)
  if (len(uriChunks) > 2 &&
      ((len(uriChunks) == 2 && len(uriChunks[1]) == 0) &&
       len(r.URL.RawQuery) == 0 )) {
    LogRequest(r,403)
    http.Error(w, "Not Acceptable", 403)
    return
  }

  var filename string
  info := Info{
    Ip: r.RemoteAddr,
    Random: rand.Int63(),
  }

  if (len(uriChunks) == 2 && len(uriChunks[1]) != 0) {
    filename = uriChunks[1]
  }
  params := parseRawQuery(r.URL.RawQuery)
  var p_ext string
  for k,v := range params {
    switch {
    case (k == "filename"):
      filename = v
    case (k == "ext"):
      if (v[0] != '.') {
        p_ext = fmt.Sprintf(".%s", v)
      } else {
        p_ext = v
      }
    case (k == "k" || k == "key" || k == "keyword"):
      info.Keywords = append(info.Keywords[:], v)
    case (k == "keys" || k == "keywords"):
      for _, key := range strings.Split(v, ",") {
        info.Keywords = append(info.Keywords[:], key)
      }
    }
  }

  if (len(filename) == 0) {
    str := GetSmallHash()
    if (len(p_ext) == 0) {
      filename = fmt.Sprintf("%s.jpg", str)
    } else {
      filename = fmt.Sprintf("%s%s", str, p_ext)
    }
  }

  exists, err := HasFileByFilename(filename)
  if (err == nil && !exists) {
    file, err := gfs.Create(filename)
    if (err != nil) {
      serverErr(w,r,err)
      return
    }
    n, err := io.Copy(file, r.Body)
    if (err != nil) {
      serverErr(w,r,err)
      return
    }
    if (n != r.ContentLength) {
      log.Printf("WARNING: [%s] content-length (%d), content written (%d)",
          filename,
          r.ContentLength,
          n)
    }
    file.Close()
  } else if (exists) {
    log.Printf("[%s] already exists", filename)
  } else {
    serverErr(w,r,err)
    return
  }

  io.WriteString(w,
      fmt.Sprintf("%s%s/f/%s\n", r.URL.Scheme, r.URL.Host, filename))

  LogRequest(r,200)
}

func routeFilesPUT(w http.ResponseWriter, r *http.Request) {
  // update the file by the name in the path and/or parameter?
  // update/add keywords from the parameters
  // look for an image in the r.Body
  LogRequest(r,200)
}

func routeFilesDELETE(w http.ResponseWriter, r *http.Request) {
  uriChunks := chunkURI(r.URL.Path)
  if ( len(uriChunks) > 2 ) {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  } else if (len(uriChunks) == 2 && len(uriChunks[1]) == 0) {
  }
  exists, err := HasFileByFilename(uriChunks[1])
  if (err != nil) {
    serverErr(w,r,err)
    return
  }
  if (exists) {
    err = gfs.Remove(uriChunks[1])
    if (err != nil) {
      serverErr(w,r,err)
      return
    }
    LogRequest(r,200)
  } else {
    LogRequest(r,404)
    http.NotFound(w,r)
  }
  // delete the name in the path and/or parameter?
}

func routeFiles(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    routeFilesGET(w,r)
  case r.Method == "PUT":
    routeFilesPUT(w,r)
  case r.Method == "POST":
    routeFilesPOST(w,r)
  case r.Method == "DELETE":
    routeFilesDELETE(w,r)
  default:
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }
}

func routeRoot(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }
  // Show a page of most recent images, and tags, and uploaders ...

  w.Header().Set("Content-Type", "text/html")
  //iter := gfs.Find(bson.M{"uploadDate": bson.M{"$gt": time.Now().Add(-time.Hour)}}).Limit(10).Iter()
  iter := gfs.Find(nil).Sort("-uploadDate").Limit(10).Iter()
  writeList(w, iter)
  LogRequest(r,200)
}

func routeAll(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  w.Header().Set("Content-Type", "text/html")

  // Show a page of all the images
  iter := gfs.Find(nil).Iter()
  writeList(w, iter)
  LogRequest(r,200)
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
  if (r.Method != "GET" ||
      len(uriChunks) > 3 ||
      (len(uriChunks) == 3 && uriChunks[2] != "r")) {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  log.Printf("K: %s (%d)", uriChunks, len(uriChunks))
  params := parseRawQuery(r.URL.RawQuery)
  log.Printf("K: params: %s", params)

  var iter *mgo.Iter
  if (len(uriChunks) == 1) {
    // show a sorted list of tag name links
    iter = gfs.Find(bson.M{"metadata": bson.M{"keywords": uriChunks[1] } }).Sort("$natural").Limit(100).Iter()
  } else if (len(uriChunks) == 2) {
    iter = gfs.Find(bson.M{"metadata": bson.M{"keywords": uriChunks[1] } }).Limit(10).Iter()
  } else if (uriChunks[2] == "r") {
    // TODO determine how to show a random image by keyword ...
    log.Println("random isn't built yet")
    LogRequest(r,404)
    return
  }

  writeList(w, iter)

  LogRequest(r,200)
}

func writeList(w http.ResponseWriter, iter *mgo.Iter) {
  var this_file File
  fmt.Fprintf(w, "<ul>\n")
  for iter.Next(&this_file) {
    log.Println(this_file.Filename)
    fmt.Fprintf(w, "<li>%s - %d</li>\n",
        linkToFile("", this_file.Filename),
        this_file.UploadDate.Year())
  }
  fmt.Fprintf(w, "</ul>\n")
}

// Show a page of all the uploader's IPs, and the images
func routeExt(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  LogRequest(r,200)
}

// Show a page of all the uploader's IPs, and the images
func routeIPs(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  LogRequest(r,200)
}

func initMongo() {
  mongo_session, err := mgo.Dial(MongoHost)
  if err != nil {
    log.Panic(err)
  }
  images_db = mongo_session.DB(MongoDB)
  if (len(MongoUsername) > 0 && len(MongoPassword) > 0) {
    err = images_db.Login(MongoUsername, MongoPassword)
    if (err != nil) {
      log.Panic(err)
    }
  }
	gfs = images_db.GridFS("fs")
}

/* Run as the image server */
func runServer(ip, port string) {
  var addr = fmt.Sprintf("%s:%s", ip, port)

  initMongo()
  defer mongo_session.Close()

  http.HandleFunc("/", routeRoot)
  http.HandleFunc("/all", routeAll)
  http.HandleFunc("/f/", routeFiles)
  http.HandleFunc("/k/", routeKeywords)
  http.HandleFunc("/ip/", routeIPs)
  http.HandleFunc("/ext/", routeExt)
  //http.HandleFunc("/md5/", routeMD5s)

  log.Printf("Serving on %s ...", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}

