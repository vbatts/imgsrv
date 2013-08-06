package main

import (
	"github.com/vbatts/imgsrv/hash"
	"github.com/vbatts/imgsrv/types"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

/* gfs is a *mgo.GridFS defined in imgsrv.go */

func GetFileByFilename(filename string) (this_file types.File, err error) {
	err = gfs.Find(bson.M{"filename": filename}).One(&this_file)
	if err != nil {
		return this_file, err
	}
	return this_file, nil
}

func GetFileRandom() (this_file types.File, err error) {
	r := hash.Rand64()
	err = gfs.Find(bson.M{"random": bson.M{"$gt": r}}).One(&this_file)
	if err != nil {
		return this_file, err
	}
	if len(this_file.Md5) == 0 {
		err = gfs.Find(bson.M{"random": bson.M{"$lt": r}}).One(&this_file)
	}
	if err != nil {
		return this_file, err
	}
	return this_file, nil
}

/* Check whether this types.File filename is on Mongo */
func HasFileByFilename(filename string) (exists bool, err error) {
	c, err := gfs.Find(bson.M{"filename": filename}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

func HasFileByMd5(md5 string) (exists bool, err error) {
	c, err := gfs.Find(bson.M{"md5": md5}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

func HasFileByKeyword(keyword string) (exists bool, err error) {
	c, err := gfs.Find(bson.M{"metadata": bson.M{"keywords": keyword}}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

func GetKeywords() (kp []types.KeywordCount, err error) {
	job := &mgo.MapReduce{
		Map: `
    function() {
        if (!this.metadata.keywords) {
          return;
        }

        for (index in this.metadata.keywords) {
          emit(this.metadata.keywords[index], 1);
        }
    }
    `,
		Reduce: `
    function(previous, current) {
      var count = 0;

      for (index in current) {
        count += current[index];
      }

      return count;
    }
    `,
	}
	if _, err := gfs.Find(nil).MapReduce(job, &kp); err != nil {
		return kp, err
	}
	return kp, nil
}
