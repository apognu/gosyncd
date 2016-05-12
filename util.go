package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

var (
	bold    = color.New(color.Bold).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	magenta = color.New(color.FgHiMagenta).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
)

func doRequest(c *cli.Context, file string, remote string, url string, body io.Reader) *http.Response {
	if !dryRun {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s%s", remote, url), body)
		resp, _ := http.DefaultClient.Do(req)

		return resp
	}
	return nil
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	io.Copy(hasher, file)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type logFunc func(remote, path string)

func logAdd(remote, path string) {
	logrus.Infof("%s: %s %s %s added", time.Now().Format("2006/02/01 15:04:05 -0700"), green("[+]"), blue(remote), bold(path))
}

func logUpdate(remote, path string) {
	logrus.Infof("%s: %s %s %s updated", time.Now().Format("2006/02/01 15:04:05 -0700"), magenta("[M]"), blue(remote), bold(path))
}

func logDelete(remote, path string) {
	logrus.Infof("%s: %s %s %s deleted", time.Now().Format("2006/02/01 15:04:05 -0700"), yellow("[-]"), blue(remote), bold(path))
}

func logWarning(remote, path, cause string, tokens ...interface{}) {
	logrus.Warnf("%s: %s %s %s skipped: %s", time.Now().Format("2006/02/01 15:04:05 -0700"), red("[!]"), blue(remote), bold(path), fmt.Sprintf(cause, tokens...))
}

func handleResponse(remote, path string, resp *http.Response, log logFunc) {
	if resp == nil || resp.StatusCode == http.StatusAccepted {
		log(remote, path)
	} else {
		logWarning(remote, path, "remote could not process file")
	}
}
