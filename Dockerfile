FROM fedora
EXPOSE 7777
ENV GOPATH /usr/local
ENV MONGODB_DATABASE filesrv
ENV MONGODB_SERVICE_HOST 127.0.0.1
ENV MONGODB_SERVICE_PORT 27017
RUN dnf install -y golang git bzr && \
	go get github.com/vbatts/imgsrv && \
	rm -rf /usr/local/pkg && \
	dnf remove -y golang git bzr
ENTRYPOINT ["/bin/sh", "/usr/local/src/github.com/vbatts/imgsrv/run.sh"]
