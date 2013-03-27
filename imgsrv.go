package main

/*
 Fun Times. This is to serve as a single utility that 
 * can be an image server, that stores into a mongo backend,
 OR
 * the client side tool that pushes/pulls images to the running server.
*/

import (
	"flag"
	"fmt"
	"labix.org/v2/mgo"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

var (
	ConfigFile = fmt.Sprintf("%s/.imgsrv.yaml", os.Getenv("HOME"))

	DefaultRunAsServer = false
	RunAsServer        = DefaultRunAsServer

	DefaultServerIP      = "0.0.0.0"
	ServerIP             = DefaultServerIP
	DefaultServerPort    = "7777"
	ServerPort           = DefaultServerPort
	DefaultMongoHost     = "localhost"
	MongoHost            = DefaultMongoHost
	DefaultMongoDB       = "filesrv"
	MongoDB              = DefaultMongoDB
	DefaultMongoUsername = ""
	MongoUsername        = DefaultMongoUsername
	DefaultMongoPassword = ""
	MongoPassword        = DefaultMongoPassword

	mongo_session *mgo.Session
	images_db     *mgo.Database
	gfs           *mgo.GridFS

	DefaultRemoteHost = ""
	RemoteHost        = DefaultRemoteHost
	PutFile           = ""
	FetchUrl          = ""
	FileKeywords      = ""
)

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		// What to do with these floating args ...
		log.Printf("%s", arg)
	}

	// loads either default or flag specified config
	// to override variables
	loadConfiguration(ConfigFile)

	if len(FetchUrl) > 0 {
		file, err := FetchFileFromURL(FetchUrl)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(file)
	} else if RunAsServer {
		log.Printf("%s", ServerIP)
		runServer(ServerIP, ServerPort)
	} else {
		if len(RemoteHost) == 0 {
			log.Println("Please provide a remotehost!")
			return
		}
		if len(PutFile) == 0 { //&& len(flag.Args()) == 0) {
			log.Println("Please provide files to be uploaded!")
			return
		}
		_, basename := filepath.Split(PutFile)
		queryParams := "?filename=" + basename
		if len(FileKeywords) > 0 {
			queryParams = queryParams + "&keywords=" + FileKeywords
		} else {
			log.Println("WARN: you didn't provide any keywords :-(")
		}
		url, err := url.Parse(RemoteHost + "/f/" + queryParams)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("POSTing: %s\n", url.String())
		url_path, err := PutFileFromPath(url.String(), PutFile)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("New Image!: %s%s\n", RemoteHost, url_path)
	}
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
	flag.StringVar(&MongoUsername,
		"mongo-username",
		MongoUsername,
		"Mongo username to auth with (if needed) ('mongousername' in the config)")
	flag.StringVar(&MongoPassword,
		"mongo-password",
		MongoPassword,
		"Mongo password to auth with (if needed) ('mongopassword' in the config)")

	/* Client-side */
	flag.StringVar(&FetchUrl,
		"fetch",
		FetchUrl,
		"Just fetch the file from this url")

	flag.StringVar(&RemoteHost,
		"remotehost",
		RemoteHost,
		"Remote host to get/put files on ('remotehost' in the config)")
	flag.StringVar(&PutFile,
		"put",
		PutFile,
		"Put file on remote server (needs -remotehost)")
	flag.StringVar(&FileKeywords,
		"keywords",
		FileKeywords,
		"Keywords to associate with file. (comma delimited) (needs -put)")

}

func loadConfiguration(filename string) (c Config) {
	//log.Printf("Attempting to load config file: %s", filename)
	c, err := ReadConfigFile(filename)
	if err != nil {
		//log.Println(err)
		return Config{}
	}

	cRunAsServer := c.GetBool("server")
	cServerIp := c.GetString("ip")
	cServerPort := c.GetString("port")
	cMongoHost := c.GetString("mongohost")
	cMongoDB := c.GetString("mongodb")
	cMongoUsername := c.GetString("mongousername")
	cMongoPassword := c.GetString("mongopassword")
	cRemoteHost := c.GetString("remotehost")

	// Only set variables from config file,
	// if they weren't passed as flags
	if DefaultRunAsServer == RunAsServer && cRunAsServer {
		RunAsServer = cRunAsServer
	}
	if DefaultServerIP == ServerIP && len(cServerIp) > 0 {
		ServerIP = cServerIp
	}
	if DefaultServerPort == ServerPort && len(cServerPort) > 0 {
		ServerPort = cServerPort
	}
	if DefaultMongoHost == MongoHost && len(cMongoHost) > 0 {
		MongoHost = cMongoHost
	}
	if DefaultMongoDB == MongoDB && len(cMongoDB) > 0 {
		MongoDB = cMongoDB
	}
	if DefaultMongoUsername == MongoUsername && len(cMongoUsername) > 0 {
		MongoUsername = cMongoUsername
	}
	if DefaultMongoPassword == MongoPassword && len(cMongoPassword) > 0 {
		MongoPassword = cMongoPassword
	}
	if DefaultRemoteHost == RemoteHost && len(cRemoteHost) > 0 {
		RemoteHost = cRemoteHost
	}

	return c
}
