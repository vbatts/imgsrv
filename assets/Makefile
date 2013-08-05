
PKG := assets

all: \
	bootstrap.css.go \
	jquery.js.go \
	jqud.js.go \
	bootstrap.js.go

bootstrap.css.go: bootstrap.css 
	go-bindata -pkg=$(PKG) -func="BootstrapCss" -out=$@ $^

jquery.js.go: jquery-2.0.3.js
	go-bindata -pkg=$(PKG) -func="JqueryJs" -out=$@ $^

jqud.js.go: jqud.js
	go-bindata -pkg=$(PKG) -func="TagCloudJs" -out=$@ $^

bootstrap.js.go: bootstrap.js
	go-bindata -pkg=$(PKG) -func="BootstrapJs" -out=$@ $^

