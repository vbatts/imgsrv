package main

import (
  "crypto/md5"
  "fmt"
  "hash/adler32"
  "io"
  "math/rand"
  "time"
)

func Rand64() int64 {
  return rand.Int63()
}

/* Convinience method for getting md5 sum of a string */
func GetMd5FromString(blob string) (sum []byte) {
  h := md5.New()
  defer h.Reset()
  io.WriteString(h,blob)
  return h.Sum(nil)
}

/* Convinience method for getting md5 sum of some bytes */
func GetMd5FromBytes(blob []byte) (sum []byte) {
  h := md5.New()
  defer h.Reset()
  h.Write(blob)
  return h.Sum(nil)
}

/* get a small, decently unique hash */
func GetSmallHash() (small_hash string) {
  h := adler32.New()
  io.WriteString(h, fmt.Sprintf("%d", time.Now().Unix()))
  return fmt.Sprintf("%X", h.Sum(nil))
}

