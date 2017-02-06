FROM fedora
RUN dnf install -y golang
RUN mkdir -p /usr/local/src/github.com/vbatts/
ENV GOPATH=/usr/local
ADD ./ /usr/local/src/github.com/vbatts/imgsrv/
RUN go install github.com/vbatts/imgsrv
EXPOSE 7777
ENTRYPOINT ["/usr/local/bin/imgsrv"]
