package mongo

import (
	"strings"

	"github.com/vbatts/imgsrv/dbutil"
	"github.com/vbatts/imgsrv/types"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func init() {
	dbutil.Handles["mongo"] = &mongoHandle{}
}

const defaultDbName = "filesrv"

type dbConfig struct {
	Seed   string // mongo host seed to Dial into
	User   string // mongo credentials, if needed
	Pass   string // mongo credentials, if needed
	DbName string // mongo database name, if needed
}

type mongoHandle struct {
	config  dbConfig
	Session *mgo.Session
	FileDb  *mgo.Database
	Gfs     *mgo.GridFS
}

func (h *mongoHandle) Init(config interface{}) (err error) {
	h.config = config.(dbConfig)

	h.Session, err = mgo.Dial(h.config.Seed)
	if err != nil {
		return err
	}

	if len(h.config.DbName) > 0 {
		h.FileDb = h.Session.DB(h.config.DbName)
	} else {
		h.FileDb = h.Session.DB(defaultDbName)
	}

	if len(h.config.User) > 0 && len(h.config.Pass) > 0 {
		err = h.FileDb.Login(h.config.User, h.config.Pass)
		if err != nil {
			return err
		}
	}
	h.Gfs = h.FileDb.GridFS("fs")
	return nil
}

func (h mongoHandle) Close() error {
	h.Session.Close()
	return nil
}

// pass through for GridFs
func (h mongoHandle) Open(filename string) (file *mgo.GridFile, err error) {
	return h.Gfs.Open(strings.ToLower(filename))
}

// pass through for GridFs
func (h mongoHandle) Create(filename string) (file *mgo.GridFile, err error) {
	return h.Gfs.Create(strings.ToLower(filename))
}

// pass through for GridFs
func (h mongoHandle) Remove(filename string) (err error) {
	return h.Gfs.Remove(strings.ToLower(filename))
}

// Find files by their MD5 checksum
func (h mongoHandle) FindFilesByMd5(md5 string) (files []types.File, err error) {
	err = h.Gfs.Find(bson.M{"md5": md5}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

// match for file name
// XXX this is not used
func (h mongoHandle) FindFilesByName(filename string) (files []types.File, err error) {
	err = h.Gfs.Find(bson.M{"filename": filename}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

// Case-insensitive pattern match for file name
func (h mongoHandle) FindFilesByPatt(filenamePat string) (files []types.File, err error) {
	err = h.Gfs.Find(bson.M{"filename": bson.M{"$regex": filenamePat, "$options": "i"}}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

// Case-insensitive pattern match for file name
func (h mongoHandle) FindFilesByKeyword(keyword string) (files []types.File, err error) {
	err = h.Gfs.Find(bson.M{"metadata.keywords": strings.ToLower(keyword)}).Sort("-metadata.timestamp").All(&files)
	return files, err
}

// Get all the files.
// Pass -1 for all files.
func (h mongoHandle) GetFiles(limit int) (files []types.File, err error) {
	//files = []types.File{}
	if limit == -1 {
		err = h.Gfs.Find(nil).Sort("-metadata.timestamp").All(&files)
	} else {
		err = h.Gfs.Find(nil).Sort("-metadata.timestamp").Limit(limit).All(&files)
	}
	return files, err
}

// Count the filename matches
func (h mongoHandle) CountFiles(filename string) (count int, err error) {
	query := h.Gfs.Find(bson.M{"filename": strings.ToLower(filename)})
	return query.Count()
}

// Get one file back, by searching by file name
func (h mongoHandle) GetFileByFilename(filename string) (thisFile types.File, err error) {
	err = h.Gfs.Find(bson.M{"filename": strings.ToLower(filename)}).One(&thisFile)
	if err != nil {
		return thisFile, err
	}
	return thisFile, nil
}

// Check whether this types.File filename is on Mongo
func (h mongoHandle) HasFileByFilename(filename string) (exists bool, err error) {
	c, err := h.CountFiles(filename)
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

// XXX this is not used
func (h mongoHandle) HasFileByMd5(md5 string) (exists bool, err error) {
	c, err := h.Gfs.Find(bson.M{"md5": md5}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

// XXX this is not used
func (h mongoHandle) HasFileByKeyword(keyword string) (exists bool, err error) {
	c, err := h.Gfs.Find(bson.M{"metadata": bson.M{"keywords": strings.ToLower(keyword)}}).Count()
	if err != nil {
		return false, err
	}
	exists = (c > 0)
	return exists, nil
}

// get a list of file extensions and their frequency count
func (h mongoHandle) GetExtensions() (kp []types.IdCount, err error) {
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
	if _, err := h.Gfs.Find(nil).MapReduce(job, &kp); err != nil {
		return kp, err
	}
	// Less than effecient, but cleanest place to put this
	for i := range kp {
		kp[i].Root = "ext" // for extension. Maps to /ext/
	}
	return kp, nil
}

// get a list of keywords and their frequency count
func (h mongoHandle) GetKeywords() (kp []types.IdCount, err error) {
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
	if _, err := h.Gfs.Find(nil).MapReduce(job, &kp); err != nil {
		return kp, err
	}
	// Less than effecient, but cleanest place to put this
	for i := range kp {
		kp[i].Root = "k" // for keyword. Maps to /k/
	}
	return kp, nil
}
