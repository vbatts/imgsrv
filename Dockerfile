FROM fedora
ENV GOPATH /usr/local
EXPOSE 7777
RUN dnf install -y golang git && \
	go get github.com/vbatts/imgsrv && \
	rm -rf /usr/local/src /usr/local/pkg && \
	dnf remove -y golang git
ENTRYPOINT ["/usr/local/bin/imgsrv"]
