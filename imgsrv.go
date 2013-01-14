package main

/*
 Fun Times. This is to serve as a single utility that 
 * can be an image server, that stores into a mongo backend,
 OR
 * the client side tool that pushes/pulls images to the running server.
*/

import (
  "log"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  goconf "git.corp.redhat.com/srv/git/dev/towers/engineering/go/conf.git"
  "net/http"
  "flag"
  "fmt"
  "path/filepath"
  "os"
  "strings"
  "time"
  //"errors"
  "mime"

  "crypto/md5"
  "hash/adler32"
  "io"
)

var (
  ConfigFile   = fmt.Sprintf("%s/.imgsrv.conf", os.Getenv("HOME"))

  DefaultRunAsServer    = false
  RunAsServer           = DefaultRunAsServer
  DefaultServerIP       = "0.0.0.0"
  ServerIP              = DefaultServerIP
  DefaultServerPort     = "7777"
  ServerPort            = DefaultServerPort
  DefaultMongoHost      = "localhost"
  MongoHost             = DefaultMongoHost
  DefaultMongoDB        = "images"
  MongoDB               = DefaultMongoDB
  MongoCollectionData   = "data"
  MongoCollectionImages = "data"

  mongo_session *mgo.Session
  images_db *mgo.Database
  gfs *mgo.GridFS

  DefaultRemoteHost     = ""
  RemoteHost            = DefaultRemoteHost
  GetFile               = ""
  PutFile               = ""
  FileKeywords          = ""
)

type Info struct {
  Keywords []string // tags
  Ip string // who uploaded it
}

type ImgFile struct {
  Metadata Info ",omitempty"
  Md5 string
  ChunkSize int
  UploadDate time.Time
  Length int64
  Filename string ",omitempty"
  ContentType string "contentType,omitempty"
}

/* Check whether this Image filename is on Mongo */
func hasImageByFilename(filename string) (exists bool, err error) {
  c, err := gfs.Find(bson.M{"filename": filename}).Count()
  if (err != nil) {
    return false, err
  }
  exists = (c > 0)
  return exists, nil
}

func getImage(filename string) (img ImgFile, err error) {
  err = gfs.Find(bson.M{"filename":filename}).One(&img)
  if (err != nil) {
    return img, err
  }
  return img, nil
}


/* return a <a href/> for a given filename 
   and root is the relavtive base of the explicit link.
*/
func linkToImage(root string, filename string) (html string) {
  return fmt.Sprintf("<a href='%s/i/%s'>%s</a>",
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

/* Convinience method for getting md5 sum of a string */
func getMd5FromString(blob string) (sum []byte) {
  h := md5.New()
  defer h.Reset()
  io.WriteString(h,blob)
  return h.Sum(nil)
}

/* Convinience method for getting md5 sum of some bytes */
func getMd5FromBytes(blob []byte) (sum []byte) {
  h := md5.New()
  defer h.Reset()
  h.Write(blob)
  return h.Sum(nil)
}

/* get a small, decently unique hash */
func getSmallHash() (small_hash string) {
  h := adler32.New()
  io.WriteString(h, fmt.Sprintf("%d", time.Now().Unix()))
  return fmt.Sprintf("%X", h.Sum(nil))
}

func returnErr(w http.ResponseWriter, r *http.Request, e error) {
      LogRequest(r,503)
      fmt.Fprintf(w,"Error fetching image: %s", e)
      http.Error(w, "Service Unavailable", 503)
      return
}

/*
  GET /i/
  GET /i/:name
*/
// Show a page of most recent images, and tags, and uploaders ...
func routeImagesGET(w http.ResponseWriter, r *http.Request) {
  uriChunks := chunkURI(r.URL.Path)
  if ( len(uriChunks) > 2 ) {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  if (len(uriChunks) == 2 && len(uriChunks[1]) > 0) {
    log.Printf("Searching for [%s] ...", uriChunks[1])
    query := gfs.Find(bson.M{"metadata": bson.M{"filename": uriChunks[1] } })

    c, err := query.Count()
    // preliminary checks, if they've passed an image name
    if (err != nil) {
      returnErr(w,r,err)
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
      returnErr(w,r,err)
      return
    }

    io.Copy(w,file) // send the contents of the file in the body

  } else {
    // TODO show a list of recent uploads? ...
  }
  LogRequest(r,200)
}

/*
  POST /i/[:name][?k=v&k=v]
*/
// Create the file by the name in the path and/or parameter?
// add keywords from the parameters
// look for an image in the r.Body
func routeImagesPOST(w http.ResponseWriter, r *http.Request) {
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
    str := getSmallHash()
    if (len(p_ext) == 0) {
      filename = fmt.Sprintf("%s.jpg", str)
    } else {
      filename = fmt.Sprintf("%s%s", str, p_ext)
    }
  }

  exists, err := hasImageByFilename(filename)
  if (err == nil && !exists) {
    file, err := gfs.Create(filename)
    if (err != nil) {
      returnErr(w,r,err)
      return
    }
    n, err := io.Copy(file, r.Body)
    if (err != nil) {
      returnErr(w,r,err)
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
    returnErr(w,r,err)
    return
  }

  io.WriteString(w,
      fmt.Sprintf("%s%s/i/%s\n", r.URL.Scheme, r.URL.Host, filename))

  LogRequest(r,200)
}

func routeImagesPUT(w http.ResponseWriter, r *http.Request) {
  // update the file by the name in the path and/or parameter?
  // update/add keywords from the parameters
  // look for an image in the r.Body
}

func routeImagesDELETE(w http.ResponseWriter, r *http.Request) {
  // delete the name in the path and/or parameter?
}

func routeImages(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "GET":
    routeImagesGET(w,r)
  case r.Method == "PUT":
    routeImagesPUT(w,r)
  case r.Method == "POST":
    routeImagesPOST(w,r)
  case r.Method == "DELETE":
    routeImagesDELETE(w,r)
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
  var img ImgFile
  iter := gfs.Find(bson.M{"uploadDate": bson.M{"$gt": time.Now().Add(-time.Hour)}}).Limit(10).Iter()
  fmt.Fprintf(w, "<ul>\n")
  for iter.Next(&img) {
    log.Println(img.Filename)
    fmt.Fprintf(w, "<li>%s</li>\n", linkToImage("", img.Filename))
  }
  fmt.Fprintf(w, "</ul>\n")
}

func routeAll(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }

  w.Header().Set("Content-Type", "text/html")
  var img ImgFile

  // Show a page of all the images
  iter := gfs.Find(nil).Iter()
  fmt.Fprintf(w, "<ul>\n")
  for iter.Next(&img) {
    log.Println(img.Filename)
    fmt.Fprintf(w, "<li>%s</li>\n", linkToImage("", img.Filename))
  }
  fmt.Fprintf(w, "</ul>\n")
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

  var img ImgFile
  if (len(uriChunks) == 1) {
    // show a sorted list of tag name links
    iter := gfs.Find(bson.M{"metadata": bson.M{"keywords": uriChunks[1] } }).Sort("$natural").Limit(100).Iter()
    fmt.Fprintf(w, "<li>\n")
    for iter.Next(&img) {
      log.Println(img.Filename)
      fmt.Fprintf(w, "<ul>%s</ul>\n", linkToImage("", img.Filename))
    }
    fmt.Fprintf(w, "</li>\n")
  } else if (len(uriChunks) == 2) {
    iter := gfs.Find(bson.M{"metadata": bson.M{"keywords": uriChunks[1] } }).Limit(10).Iter()

    fmt.Fprintf(w, "<li>\n")
    for iter.Next(&img) {
      log.Println(img.Filename)
      fmt.Fprintf(w, "<ul>%s</ul>\n", linkToImage("", img.Filename))
    }
    fmt.Fprintf(w, "</li>\n")
  } else if (uriChunks[2] == "r") {
    // TODO determine how to show a random image by keyword ...
  }


}

// Show a page of all the uploader's IPs, and the images
func routeIPs(w http.ResponseWriter, r *http.Request) {
  if (r.Method != "GET") {
    LogRequest(r,404)
    http.NotFound(w,r)
    return
  }
}

func initMongo() {
  mongo_session, err := mgo.Dial(MongoHost)
  if err != nil {
    log.Panic(err)
  }
  images_db = mongo_session.DB(MongoDB)
	gfs = images_db.GridFS("fs")
}

/* Run as the image server */
func runServer(ip, port string) {
  var addr = fmt.Sprintf("%s:%s", ip, port)

  initMongo()
  defer mongo_session.Close()

  http.HandleFunc("/", routeRoot)
  http.HandleFunc("/all", routeAll)
  http.HandleFunc("/i/", routeImages)
  http.HandleFunc("/k/", routeKeywords)
  http.HandleFunc("/ip/", routeIPs)

  log.Printf("Serving on %s ...", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}

/* http://golang.org/doc/effective_go.html#init */
func init() {
  flag.StringVar(&ConfigFile,
       "config",
       ConfigFile,
       "Provide alternate configuration file")

  /* Server-side */
  flag.BoolVar(&RunAsServer,
       "server",
       RunAsServer,
       "Run as an image server (defaults to client-side)")
  flag.StringVar(&ServerIP,
       "ip",
       ServerIP,
       "IP to bind to (if running as a server)('ip' in the config)")
  flag.StringVar(&ServerPort,
       "port",
       ServerPort,
       "Port to listen on (if running as a server)('port' in the config)")

  /* MongoDB settings */
  flag.StringVar(&MongoHost,
       "mongo-host",
       MongoHost,
       "Mongo Host to connect to ('mongohost' in the config)")
  flag.StringVar(&MongoDB,
       "mongo-db",
       MongoDB,
       "Mongo db to connect to ('mongodb' in the config)")

  /* Client-side */
  flag.StringVar(&RemoteHost,
       "remotehost",
       RemoteHost,
       "Remote host to get/put files on ('remotehost' in the config)")
  flag.StringVar(&GetFile,
       "get",
       GetFile,
       "Fetch file on remote server (needs -remotehost)")
  flag.StringVar(&PutFile,
       "put",
       PutFile,
       "Put file on remote server (needs -remotehost)")
  flag.StringVar(&FileKeywords,
       "keywords",
       FileKeywords,
       "Keywords to associate with file. (comma delimited) (needs -put)")

}

func loadConfiguration(filename string) (c *goconf.ConfigFile) {
  //log.Printf("Attempting to load config file: %s", filename)
  c, err := goconf.ReadConfigFile(filename)
  if (err != nil) {
    //log.Println(err)
    return goconf.NewConfigFile()
  }

  cRunAsServer, _ := c.GetBool("", "server")
  cServerIp, _ := c.GetString("", "ip")
  cServerPort, _ := c.GetString("", "port")
  cMongoHost, _ := c.GetString("", "mongohost")
  cMongoDB, _ := c.GetString("", "mongodb")
  cRemoteHost, _ := c.GetString("", "remotehost")

  // Only set variables from config file,
  // if they weren't passed as flags
  if (DefaultRunAsServer == RunAsServer && cRunAsServer) {
    RunAsServer = cRunAsServer
  }
  if (DefaultServerIP == ServerIP && len(cServerIp) > 0) {
    ServerIP = cServerIp
  }
  if (DefaultServerPort == ServerPort && len(cServerPort) > 0) {
    ServerPort = cServerPort
  }
  if (DefaultMongoHost == MongoHost && len(cMongoHost) > 0) {
    MongoHost = cMongoHost
  }
  if (DefaultMongoDB == MongoDB && len(cMongoDB) > 0) {
    MongoDB = cMongoDB
  }
  if (DefaultRemoteHost == RemoteHost && len(cRemoteHost) > 0) {
    RemoteHost = cRemoteHost
  }

  return c
}

func main() {
  flag.Parse()
  for _, arg := range flag.Args() {
    // What to do with these floating args ...
    log.Printf("%s", arg)
  }

  // loads either default or flag specified config
  // to override variables
  loadConfiguration(ConfigFile)

  if (RunAsServer) {
    runServer(ServerIP,ServerPort)
  }

  log.Printf("%s", ServerIP)
  log.Printf("%d", len(GetFile))

  //log.Printf("Good Morning!")
  //log.Printf("%x", getMd5FromString("Good Morning!"))
}

