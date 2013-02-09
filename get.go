package main

import (
  "labix.org/v2/mgo/bson"
)

/* gfs is a *mgo.GridFS defined in imgsrv.go */

func GetFileByFilename(filename string) (this_file File, err error) {
  err = gfs.Find(bson.M{"filename":filename}).One(&this_file)
  if (err != nil) {
    return this_file, err
  }
  return this_file, nil
}

func GetFileRandom() (this_file File, err error) {
  r := rand64()
  err = gfs.Find(bson.M{"random": bson.M{"$gt" : r } }).One(&this_file)
  if (err != nil) {
    return this_file, err
  }
  if (len(this_file.Md5) == 0) {
    err = gfs.Find(bson.M{"random": bson.M{"$lt" : r } }).One(&this_file)
  }
  if (err != nil) {
    return this_file, err
  }
  return this_file, nil
}

