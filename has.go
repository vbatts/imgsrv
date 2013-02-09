
package main

import (
  "labix.org/v2/mgo/bson"
)

/* Check whether this File filename is on Mongo */
func HasFileByFilename(filename string) (exists bool, err error) {
  c, err := gfs.Find(bson.M{"filename": filename}).Count()
  if (err != nil) {
    return false, err
  }
  exists = (c > 0)
  return exists, nil
}

func HasFileByMd5(md5 string) (exists bool, err error) {
  c, err := gfs.Find(bson.M{"md5": md5 }).Count()
  if (err != nil) {
    return false, err
  }
  exists = (c > 0)
  return exists, nil
}

func HasFileByKeyword(keyword string) (exists bool, err error) {
  c, err := gfs.Find(bson.M{"metadata": bson.M{"keywords": keyword} }).Count()
  if (err != nil) {
    return false, err
  }
  exists = (c > 0)
  return exists, nil
}

