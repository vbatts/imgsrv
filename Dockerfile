FROM fedora
ENV GOPATH /usr/local
EXPOSE 7777
RUN dnf install -y golang git bzr && \
	go get github.com/vbatts/imgsrv && \
	rm -rf /usr/local/src /usr/local/pkg && \
	dnf remove -y golang git bzr
ENV MONGODB_DATABASE filesrv
ENV MONGODB_SERVICE_HOST 127.0.0.1
ENV MONGODB_SERVICE_PORT 27017
ENTRYPOINT /usr/local/bin/imgsrv -server -mongo-host=$MONGODB_SERVICE_HOST:$MONGODB_SERVICE_PORT -mongo-db=$MONGODB_DATABASE -mongo-username=$MONGODB_USER -mongo-password=$MONGODB_PASSWORD
