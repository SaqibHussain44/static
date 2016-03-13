package main

import (
	"fmt"
	"strings"
)

type dir struct {
	DirPath  string `yaml:"dir_path"`
	HTTPPath string `yaml:"http_path"`
}

type authedDir struct {
	DirPath   string   `yaml:"dir_path"`
	HTTPPath  string   `yaml:"http_path"`
	Usernames []string `yaml:"usernames"`
}

type user struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type config struct {
	Logging bool `yaml:"logging"`

	HTTPLAddr   string `yaml:"http_laddr"`
	HTTPSLAddr  string `yaml:"https_laddr"`
	TLSCertPath string `yaml:"tls_cert_path"`
	TLSKeyPath  string `yaml:"tls_key_path"`

	PublicDirs        []dir       `yaml:"public_dirs"`
	AuthenticatedDirs []authedDir `yaml:"authenticated_dirs"`

	Users map[string]string `yaml:"users"`
}

func (c *config) check() (err error) {
	checkHTTPPath := func(httpPath string) (err error) {
		if !strings.HasSuffix(httpPath, "/") {
			err = fmt.Errorf("dir %s not properly configured. http path should end with '/'", httpPath)
			return
		}
		return
	}
	for _, d := range c.AuthenticatedDirs {
		if err = checkHTTPPath(d.HTTPPath); err != nil {
			return
		}
	}
	for _, d := range c.PublicDirs {
		if err = checkHTTPPath(d.HTTPPath); err != nil {
			return
		}
	}
	return
}
