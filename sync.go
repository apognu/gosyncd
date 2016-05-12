package main

import (
	"bytes"
	"fmt"
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

type localFile struct {
	Mode             os.FileMode
	Checksum         string
	PresentInRemotes map[string]bool
	LastUpdate       int64
	Skip             bool
}

var dryRun = false

func sync(c *cli.Context) error {
	if c.Command.Name == "state" {
		logrus.Info("synchronization dry run")
		dryRun = true
	}

	// Mapping all local files to be compared to remote manifest
	lDir := make(map[string]*localFile)
	for _, dir := range config.Directories {
		if _, err := os.Stat(dir.Name); os.IsNotExist(err) {
			logWarning("local", dir.Name, "directory does not exist")
			continue
		}

		filepath.Walk(dir.Name, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			chks := ""
			skip := false
			for _, excl := range dir.Exclude {
				if strings.HasPrefix(path, excl) {
					skip = true
					break
				}
			}
			if !skip {
				chks, err = fileChecksum(path)
				if err != nil {
					logWarning("local", path, err.Error())
					skip = true
				}
			}

			lDir[path] = &localFile{
				Mode:             info.Mode(),
				Checksum:         chks,
				PresentInRemotes: make(map[string]bool),
				LastUpdate:       info.ModTime().Unix(),
				Skip:             skip,
			}

			return nil
		})
	}

	logrus.Debugf("mapped %d local files in %d directories", len(lDir), len(config.Directories))

	for _, remote := range config.Remotes {
		resp, err := http.Get(fmt.Sprintf("https://%s/manifest", remote))
		if err != nil {
			logWarning(remote, "*", "could not reach remote host: %s", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			logWarning(remote, "*", "issue on remote side, skipping host")
			return nil
		}

		defer resp.Body.Close()

		rDir := directory.Directory{}
		rDirBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logWarning(remote, "*", "could not read response from remote host: %s", err)
			continue
		}

		proto.Unmarshal(rDirBytes, &rDir)

		logrus.Debugf("received remote manifest from '%s' with %d files", remote, len(rDir.GetFiles()))

		for _, rFile := range rDir.GetFiles() {
			// Remote file is still present on local host
			if _, ok := lDir[rFile.GetPath()]; ok {
				if lDir[rFile.GetPath()].Skip {
					logrus.Debugf("skipping %s for update", rFile.GetPath())
					continue
				}

				lDir[rFile.GetPath()].PresentInRemotes[remote] = true
			} else {
				processDeleted(c, remote, rFile)
				continue
			}

			if lDir[rFile.GetPath()].Checksum != rFile.GetChecksum() {
				logrus.Debugf("checksum mismatch on '%s' for %s", remote, rFile.GetPath())
				if rFile.GetLastUpdate() > lDir[rFile.GetPath()].LastUpdate {
					logWarning(remote, rFile.GetPath(), "local file is older than remote counterpart")
					continue
				}

				processUpdated(c, remote, lDir, rFile)
				continue
			}

			if uint32(lDir[rFile.GetPath()].Mode) != rFile.GetMode() {
				logrus.Debugf("mode mismatch on '%s' for %s", remote, rFile.GetPath())
				processUpdated(c, remote, lDir, rFile)
			}
		}

		processCreated(c, remote, lDir)
	}

	return nil
}

func processUpdated(c *cli.Context, remote string, lDir map[string]*localFile, rFile *directory.Directory_File) {
	lFile, err := os.Open(rFile.GetPath())
	if err != nil {
		logWarning(rFile.GetPath(), "could not open local file: %s", err.Error())
		return
	}
	p, err := ioutil.ReadAll(lFile)
	if err != nil {
		logWarning(rFile.GetPath(), "could not read local file: %s", err.Error())
		return
	}

	info, err := lFile.Stat()
	if err != nil {
		logWarning(rFile.GetPath(), "could not stat local file: %s", err.Error())
		return
	}

	update := file.File_UPDATE
	file := file.File{
		Path:    proto.String(rFile.GetPath()),
		Mode:    proto.Uint32(uint32(info.Mode())),
		Payload: p,
		Action:  &update,
	}

	filePb, err := proto.Marshal(&file)
	if err != nil {
		logWarning(rFile.GetPath(), "could marshal file to transmit: %s", err.Error())
		return
	}

	logrus.Debugf("sending update to '%s' for %s", remote, rFile.GetPath())

	resp := doRequest(c, file.GetPath(), remote, "/update", bytes.NewReader(filePb))
	handleResponse(remote, file.GetPath(), resp, logUpdate)
}

func processCreated(c *cli.Context, remote string, lDir map[string]*localFile) {
	for path, f := range lDir {
		logrus.Debugf("considering %s on %s for creation", path, remote)
		if f.Skip {
			logrus.Debugf("skipping %s for creation", path)
			continue
		}

		if !f.PresentInRemotes[remote] {
			lFile, err := os.Open(path)
			if err != nil {
				logWarning(path, "could not open local file: %s", err.Error())
				continue
			}

			p, err := ioutil.ReadAll(lFile)
			if err != nil {
				logWarning(path, "could not read local file: %s", err.Error())
				continue
			}

			info, err := lFile.Stat()
			if err != nil {
				logWarning(path, "could not stat local file: %s", err.Error())
				continue
			}

			add := file.File_ADD
			file := file.File{
				Path:    proto.String(path),
				Mode:    proto.Uint32(uint32(info.Mode())),
				Payload: p,
				Action:  &add,
			}

			filePb, err := proto.Marshal(&file)
			if err != nil {
				logWarning(path, "could not marshall file to transmit: %s", err.Error())
			}

			logrus.Debugf("sending create to '%s' for %s", remote, path)

			resp := doRequest(c, file.GetPath(), remote, "/update", bytes.NewReader(filePb))
			handleResponse(remote, file.GetPath(), resp, logAdd)
		}
	}
}

func processDeleted(c *cli.Context, remote string, rFile *directory.Directory_File) {
	delete := file.File_DELETE
	file := file.File{
		Path:    proto.String(rFile.GetPath()),
		Mode:    proto.Uint32(0),
		Payload: []byte(""),
		Action:  &delete,
	}

	filePb, err := proto.Marshal(&file)
	if err != nil {
		logWarning(rFile.GetPath(), "could not marshall file to transmit: %s", err.Error())
		return
	}

	logrus.Debugf("sending delete to '%s' for %s", remote, rFile.GetPath())

	resp := doRequest(c, file.GetPath(), remote, "/update", bytes.NewReader(filePb))
	handleResponse(remote, file.GetPath(), resp, logDelete)
}
