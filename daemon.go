package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/gosyncd/directory"
	"github.com/apognu/gosyncd/file"
	"github.com/codegangsta/cli"
	"github.com/gogo/protobuf/proto"
)

func getManifest(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("manifest requested by %s", r.RemoteAddr)

	dirs := directory.Directory{}
	errored := false
	for _, dir := range config.Directories {
		if _, err := os.Stat(dir.Name); os.IsNotExist(err) {
			errored = true
			logWarning("local", dir.Name, "directory does not exist")
			continue
		}

		filepath.Walk(dir.Name, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			chks, err := fileChecksum(path)
			if err != nil {
				return err
			}

			dirs.Files = append(dirs.Files, &directory.Directory_File{
				Path:       proto.String(path),
				Mode:       proto.Uint32(uint32(info.Mode())),
				Checksum:   proto.String(chks),
				LastUpdate: proto.Int64(info.ModTime().Unix()),
			})

			return nil
		})
	}

	logrus.Debugf("mapped %d files in %d directories", len(dirs.GetFiles()), len(config.Directories))

	if errored {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pb, _ := proto.Marshal(&dirs)
	w.Write(pb)
}

func daemon(c *cli.Context) error {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		filePb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logrus.Errorf("could not read message: %s", err)
		}
		var f file.File
		err = proto.Unmarshal(filePb, &f)
		if err != nil {
			logrus.Errorf("could not unmarshal message: %s", err)
		}

		found := false
		for _, d := range directories {
			if strings.HasPrefix(f.GetPath(), d) {
				found = true
			}
		}

		if !found {
			logWarning("local", f.GetPath(), "directory is not whitelisted for replication")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, dir := range config.Directories {
			for _, excl := range dir.Exclude {
				if strings.HasPrefix(f.GetPath(), excl) {
					logWarning("local", f.GetPath(), "file is locally excluded but was sent, are your configs different?")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		if f.GetAction() == file.File_ADD {
			logrus.Debugf("creation of %s requested by %s", f.GetPath(), r.RemoteAddr)
			create(w, f)
		} else if f.GetAction() == file.File_UPDATE {
			logrus.Debugf("update of %s requested by %s", f.GetPath(), r.RemoteAddr)
			update(w, f)
		} else if f.GetAction() == file.File_DELETE {
			logrus.Debugf("deletion of %s requested by %s", f.GetPath(), r.RemoteAddr)
			delete(w, f)
		}

	})

	http.HandleFunc("/manifest", getManifest)

	httpServer.ListenAndServeTLS(config.TLSCertificate, config.TLSPrivateKey)

	return nil
}

func create(w http.ResponseWriter, f file.File) {
	if err := ioutil.WriteFile(f.GetPath(), f.GetPayload(), os.FileMode(f.GetMode())); err != nil {
		logWarning("local", f.GetPath(), "could not create file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		logAdd("local", f.GetPath())
		w.WriteHeader(http.StatusAccepted)
	}
}

func update(w http.ResponseWriter, f file.File) {
	err := ioutil.WriteFile(f.GetPath(), f.GetPayload(), os.FileMode(f.GetMode()))
	if err != nil {
		logWarning("local", f.GetPath(), "could not update file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, err := os.Open(f.GetPath())
	if err != nil {
		logWarning("local", f.GetPath(), "could not open file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	info, err := file.Stat()
	if err != nil {
		logWarning("local", f.GetPath(), "could not stat file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if uint32(info.Mode()) != f.GetMode() {
		err = os.Chmod(f.GetPath(), os.FileMode(f.GetMode()))
		if err != nil {
			logWarning("local", f.GetPath(), "could not update mode: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	logUpdate("local", f.GetPath())
	w.WriteHeader(http.StatusAccepted)
}

func delete(w http.ResponseWriter, f file.File) {
	if err := os.Remove(f.GetPath()); err != nil {
		logWarning("local", f.GetPath(), "could not delete file: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		logDelete("local", f.GetPath())
		w.WriteHeader(http.StatusAccepted)
	}
}
