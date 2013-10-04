package dbutil

import (
	"github.com/vbatts/imgsrv/hash"
	"github.com/vbatts/imgsrv/types"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Util struct {
	Gfs *mgo.GridFS
}

/*
pass through for GridFs
*/
func (u Util) Open(filename string) (file *mgo.GridFile, err error) {
	return u.Gfs.Open(strings.ToLower(filename))
}

/*
pass through for GridFs
*/
func (u Util) Create(filename string) (file *mgo.GridFile, err error) {
	return u.Gfs.Create(strings.ToLower(filename))
}

/*
pass through for GridFs
*/
func (u Util) Remove(filename string) (err error) {
	return u.Gfs.Remove(strings.ToLower(filename))
}

/*
Find files by their MD5 checksum
*/
func (u Util) FindFilesByMd5(md5 string) (files []types.File, err error) {
	err = u.Gfs.Find(bson.M{"md5": md5}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

/*
match for file name
*/
func (u Util) FindFilesByName(filename string) (files []types.File, err error) {
	err = u.Gfs.Find(bson.M{"filename": filename}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

/*
Case-insensitive pattern match for file name
*/
func (u Util) FindFilesByPatt(filename_pat string) (files []types.File, err error) {
	err = u.Gfs.Find(bson.M{"filename": bson.M{"$regex": filename_pat, "$options": "i"}}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

/*
Case-insensitive pattern match for file name
*/
func (u Util) FindFilesByKeyword(keyword string) (files []types.File, err error) {
	err = u.Gfs.Find(bson.M{"metadata.keywords": strings.ToLower(keyword)}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

/*
Get all the files.

pass -1 for all files
*/
func (u Util) GetFiles(limit int) (files []types.File, err error) {
	if limit == -1 {
		err = u.Gfs.Find(nil).Sort("-metadata.timestamp").Limit(limit).All(&files)
	} else {
		err = u.Gfs.Find(nil).Sort("-metadata.timestamp").All(&files)
	}
	return files, err
}

/*
Count the filename matches
*/
func (u Util) CountFiles(filename string) (count int, err error) {
	query := u.Gfs.Find(bson.M{"filename": strings.ToLower(filename)})
	return query.Count()
}

/*
Get one file back, by searching by file name
*/
func (u Util) GetFileByFilename(filename string) (this_file types.File, err error) {
	err = u.Gfs.Find(bson.M{"filename": strings.ToLower(filename)}).One(&this_file)
	if err != nil {
		return this_file, err
	}
	return this_file, nil
}

func (u Util) GetFileRandom() (this_file types.File, err error) {
	r := hash.Rand64()
	err = u.Gfs.Find(bson.M{"random": bson.M{"$gt": r}}).One(&this_file)
	if err != nil {
		return this_file, err
	}
	if len(this_file.Md5) == 0 {
		err = u.Gfs.Find(bson.M{"random": bson.M{"$lt": r}}).One(&this_file)
	}
	if err != nil {
		return this_file, err
	}
	return this_file, nil
}

/*
Check whether this types.File filename is on Mongo
*/
func (u Util) HasFileByFilename(filename string) (exists bool, err error) {
	c, err := u.CountFiles(filename)
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

func (u Util) HasFileByMd5(md5 string) (exists bool, err error) {
	c, err := u.Gfs.Find(bson.M{"md5": md5}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

func (u Util) HasFileByKeyword(keyword string) (exists bool, err error) {
	c, err := u.Gfs.Find(bson.M{"metadata": bson.M{"keywords": strings.ToLower(keyword)}}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

/*
get a list of file extensions and their frequency count
*/
func (u Util) GetExtensions() (kp []types.IdCount, err error) {
	job := &mgo.MapReduce{
		Map: `
    function() {
        if (!this.filename) {
          return;
        }

        s = this.filename.split(".")
        ext = s[s.length - 1] // get the last segment of the split
        emit(ext,1);
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
	if _, err := u.Gfs.Find(nil).MapReduce(job, &kp); err != nil {
		return kp, err
	}
	// Less than effecient, but cleanest place to put this
	for i := range kp {
		kp[i].Root = "ext" // for extension. Maps to /ext/
	}
	return kp, nil
}

/*
get a list of keywords and their frequency count
*/
func (u Util) GetKeywords() (kp []types.IdCount, err error) {
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
	if _, err := u.Gfs.Find(nil).MapReduce(job, &kp); err != nil {
		return kp, err
	}
	// Less than effecient, but cleanest place to put this
	for i := range kp {
		kp[i].Root = "k" // for keyword. Maps to /k/
	}
	return kp, nil
}
