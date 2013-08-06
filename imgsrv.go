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
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/vbatts/imgsrv/client"
	"github.com/vbatts/imgsrv/config"
	"github.com/vbatts/imgsrv/util"
)

var (
	ConfigFile = fmt.Sprintf("%s/.imgsrv.yaml", os.Getenv("HOME"))

	DefaultConfig = &config.Config{
		Server:        false,
		Ip:            "0.0.0.0",
		Port:          "7777",
		MongoHost:     "localhost",
		MongoDB:       "filesrv",
		MongoUsername: "",
		MongoPassword: "",
		RemoteHost:    "",
	}

	PutFile      = ""
	FetchUrl     = ""
	FileKeywords = ""
)

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		// TODO What to do with these floating args ...
		//      Assume they're files and upload them?
		log.Printf("%s", arg)
	}

	// loads either default or flag specified config
	// to override variables
	if c, err := config.ReadConfigFile(ConfigFile); err == nil {
		DefaultConfig.Merge(c)
	}

	if DefaultConfig.Server {
		// Run the server!

		runServer(DefaultConfig)

	} else if len(FetchUrl) > 0 {
		// not sure that this ought to be exposed in the client tool

		file, err := util.FetchFileFromURL(FetchUrl)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(file)

	} else {
		// we're pushing up a file

		if len(DefaultConfig.RemoteHost) == 0 {
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
		url, err := url.Parse(DefaultConfig.RemoteHost + "/f/" + queryParams)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("POSTing: %s\n", url.String())
		url_path, err := client.PutFileFromPath(url.String(), PutFile)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("New Image!: %s%s\n", DefaultConfig.RemoteHost, url_path)
	}
}

/*
http://golang.org/doc/effective_go.html#init

TODO refactor flags and config, to assign, instead of pass by reference.
http://play.golang.org/p/XhGqn-MOjL
*/
func init() {
	flag.StringVar(&ConfigFile,
		"config",
		ConfigFile,
		"Provide alternate configuration file")

	/* Server-side */
	flag.BoolVar(&DefaultConfig.Server,
		"server",
		DefaultConfig.Server,
		"Run as an image server (defaults to client-side)")
	flag.StringVar(&DefaultConfig.Ip,
		"ip",
		DefaultConfig.Ip,
		"IP to bind to (if running as a server)('ip' in the config)")
	flag.StringVar(&DefaultConfig.Port,
		"port",
		DefaultConfig.Port,
		"Port to listen on (if running as a server)('port' in the config)")

	/* MongoDB settings */
	flag.StringVar(&DefaultConfig.MongoHost,
		"mongo-host",
		DefaultConfig.MongoHost,
		"Mongo Host to connect to ('mongohost' in the config)")
	flag.StringVar(&DefaultConfig.MongoDB,
		"mongo-db",
		DefaultConfig.MongoDB,
		"Mongo db to connect to ('mongodb' in the config)")
	flag.StringVar(&DefaultConfig.MongoUsername,
		"mongo-username",
		DefaultConfig.MongoUsername,
		"Mongo username to auth with (if needed) ('mongousername' in the config)")
	flag.StringVar(&DefaultConfig.MongoPassword,
		"mongo-password",
		DefaultConfig.MongoPassword,
		"Mongo password to auth with (if needed) ('mongopassword' in the config)")

	/* Client-side */
	flag.StringVar(&FetchUrl,
		"fetch",
		FetchUrl,
		"Just fetch the file from this url")

	flag.StringVar(&DefaultConfig.RemoteHost,
		"remotehost",
		DefaultConfig.RemoteHost,
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
