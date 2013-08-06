Overview
========

Initially, I just wanted a server that I could upload images to, and quickly
serve back. While it's mostly that, it is the client and server side tooling.

The files are stored in mongoDB, using its GridFS.

Usage
-----

Server side
For a basic start, make sure mongo is running on the localhost, and run:
  ./imgsrv -server
  2013/02/12 13:03:37 0.0.0.0
  2013/02/12 13:03:37 Serving on 0.0.0.0:7777 ...

For something a bit more complicated, like an openshift diy-0.1 cartridge, 
set your .openshift/action_hooks/start to:

	nohup ${OPENSHIFT_REPO_DIR}/bin/server \
	  -server \
	  -ip ${OPENSHIFT_INTERNAL_IP} \
	  -port 8080 \
	  -mongo-host ${OPENSHIFT_NOSQL_DB_URL} \
	  >> ${OPENSHIFT_LOG_DIR}/server.log 2>&1 &


Client side:
Either pass the -remotehost flag pointing to your server instance

	imgsrv -remotehost http://hurp.til.derp.com:7777 -put ./lolz.gif -keywords 'cats,lols'


or setup your ~/.imgsrv.yaml, with a the value 'remotehost', like so:

	---
	remotehost: http://hurp.til.derp.com:7777

then you can drop that flag out, for quicker lolzing:

	imgsrv -put ./lolz.gif -keywords 'cats,lols'
	2013/02/12 13:00:28 POSTing: http://hurp.til.derp.com:7777/f/?filename=lolz.gif&keywords=cats,lols
	2013/02/12 13:00:29 New Image!: http://hurp.til.derp.com:7777/f/lolz.gif


Dependencies
------------

	go get launchpad.net/goyaml
	go get labix.org/v2/mgo

and put this imgsrv in your GOPATH,
since it references itself.

Building
--------

Either 

	git clone git://<host>/imgsrv.git
	cd imgsrv
	go build
	./imgsrv

or

	go get github.com/vbatts/imgsrv
	imgsrv


