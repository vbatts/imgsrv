FROM fedora
ENV GOPATH /usr/local
EXPOSE 7777
RUN dnf install -y golang git bzr && \
	go get github.com/vbatts/imgsrv && \
	rm -rf /usr/local/src /usr/local/pkg && \
	dnf remove -y golang git bzr
ENTRYPOINT ["/usr/local/bin/imgsrv"]
