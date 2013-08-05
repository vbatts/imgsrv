
PKG := assets

bootstrap.css.go: bootstrap.css 
	go-bindata -pkg=$(PKG) -func="BootstrapCss" -out=$@ $^

