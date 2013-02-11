package main

import "time"

type Info struct {
  Keywords []string // tags
  Ip string // who uploaded it
  Random int64
  TimeStamp time.Time "timestamp,omitempty"
}

type File struct {
  Metadata Info ",omitempty"
  Md5 string
  ChunkSize int
  UploadDate time.Time
  Length int64
  Filename string ",omitempty"
  ContentType string "contentType,omitempty"
}
