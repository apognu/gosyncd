package main

import (
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/hashicorp/hcl"
)

var (
	config      *Config
	directories = make([]string, 0)
)

type Config struct {
	BindAddress      string            `hcl:"bind_address"`
	Remotes          []string          `hcl:"remotes"`
	TLSCACertificate string            `hcl:"tls_ca_certificate"`
	TLSCertificate   string            `hcl:"tls_certificate"`
	TLSPrivateKey    string            `hcl:"tls_private_key"`
	Directories      []ConfigDirectory `hcl:"directory"`
}

type ConfigDirectory struct {
	Name    string   `hcl:",key"`
	Exclude []string `hcl:"exclude"`
}

func setupConfig(c *cli.Context) error {
	if c.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	configFile, err := os.Open(c.GlobalString("config"))
	if err != nil {
		logrus.Fatalf("could not open config file: %s", err)
	}
	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		logrus.Fatalf("could not open config file: %s", err)
	}

	err = hcl.Unmarshal(configBytes, &config)
	if err != nil {
		logrus.Fatalf("could not unmarshal config file: %s", err)
	}

	// Loop over all remotes and all network intefaces to remove self
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for idx, remote := range config.Remotes {
			ip, _, err := net.SplitHostPort(remote)
			if err != nil {
				logrus.Fatalf("could not understand remote host: %s", remote)
			}

			for _, addr := range addrs {
				bits := strings.Split(addr.String(), "/")
				if ip == bits[0] {
					logrus.Debugf("blacklisting remote '%s' as self", ip)
					config.Remotes = append(config.Remotes[:idx], config.Remotes[idx+1:]...)
				}
			}
		}
	}

	for _, dir := range config.Directories {
		directories = append(directories, dir.Name)
	}

	initTLS(c)

	return nil
}
