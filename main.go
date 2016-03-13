package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var f struct {
	config    string
	genConfig bool
}

func init() {
	flag.StringVar(&f.config, "config", "", "path to configuration file")
	flag.BoolVar(&f.genConfig, "gen-config", false, "generate example config file and print to stdout")
}

func genConfig() {
	conf := new(config)
	conf.HTTPLAddr = ":80"
	conf.HTTPSLAddr = ":443"
	conf.TLSCertPath = "/etc/blah/example.cert"
	conf.TLSKeyPath = "/etc/blah/example.key"
	conf.Users = make(map[string]string)
	conf.Users["john"] = "efjio"
	conf.Users["huh"] = "fjweoifj"
	conf.Users["ha"] = "eioj"
	conf.PublicDirs = append(conf.PublicDirs, dir{DirPath: "/etc/www/pub1.com", HTTPPath: "/pub1/"})
	conf.PublicDirs = append(conf.PublicDirs, dir{DirPath: "/etc/www/pub2", HTTPPath: "/pub2/"})
	conf.AuthenticatedDirs = append(conf.AuthenticatedDirs, authedDir{DirPath: "/etc/www/secret", HTTPPath: "/secret/", Usernames: []string{"john", "ha"}})
	if bytes, err := yaml.Marshal(conf); err != nil {
		log.Fatalf("marshalling example config error: %s", err.Error())
	} else {
		if _, err = os.Stdout.Write(bytes); err != nil {
			log.Fatalf("writing example config to stdout error: %s", err.Error())
		}
	}
}

func main() {
	flag.Parse()

	if f.genConfig {
		genConfig()
		return
	}

	if f.config == "" {
		flag.PrintDefaults()
		return
	}

	var file *os.File
	var err error
	if file, err = os.Open(f.config); err != nil {
		log.Fatalf("opening config file %s error: %s\n", f.config, err.Error())
	}
	var bytes []byte
	if bytes, err = ioutil.ReadAll(file); err != nil {
		log.Fatalf("reading config file %s error: %s\n", f.config, err.Error())
	}
	conf := new(config)
	if err = yaml.Unmarshal(bytes, conf); err != nil {
		log.Fatalf("decoding config file %s error: %s\n", f.config, err.Error())
	}

	if err = conf.check(); err != nil {
		log.Fatalf("invalide config file %s: %s\n", f.config, err.Error())
	}

	serve(conf)

}
