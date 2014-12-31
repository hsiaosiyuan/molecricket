package proxy

import (
	"http/auth"
	"flag"
	"os"
	"log"
	"errors"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	Port      string			`json:"Port"`
	Users     []auth.User		`json:"Users"`
	Resources []auth.Resource	`json:"Resources"`
}

var (
	ERR_NO_CFG = errors.New("no config")
)

func NewConfig() (c *Config, err error) {
	var (
		cfgFile   string
		f         *os.File
		fc        []byte
	)

	flag.StringVar(&cfgFile, "c", "9090", "file contains config json string")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		log.Fatal("no config")
	}

	if f, err = os.OpenFile(cfgFile, os.O_RDONLY, 0666); err != nil {
		return nil, err
	}

	defer f.Close()

	if fc, err = ioutil.ReadAll(f); err != nil {
		return nil, err
	}

	c = new(Config)
	if err = json.Unmarshal(fc, c); err != nil {
		return nil, err
	}

	return c, nil
}
