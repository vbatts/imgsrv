FROM fedora
ENV GOPATH /usr/local
EXPOSE 7777
RUN dnf install -y golang git bzr && \
	go get github.com/vbatts/imgsrv && \
	rm -rf /usr/local/src /usr/local/pkg && \
	dnf remove -y golang git bzr
ENV MONGO_DB filesrv
ENV MONGO_HOST 127.0.0.1
ENV MONGO_PORT 27017
ENTRYPOINT /usr/local/bin/imgsrv -server -mongo-host=$MONGO_HOST:$MONGO_PORT -mongo-db=$MONGO_DB -mongo-username=$MONGO_USER -mongo-password=$MONGO_PASSWORD
