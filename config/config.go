package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"

	util "github.com/gourytch/gowowuction/util"
)

const SLASH = filepath.Separator


type Config struct {
	APIKey                         string   `json:"apikey"`
	RealmsList                     []string `json:"realms"`
	LocalesList                    []string `json:"locales"`
	DownloadDirectory              string   `json:"download_dir"`
	TempDirectory                  string   `json:"temp_dir"`
	ResultDirectory                string   `json:"result_dir"`
}

func defaultConfig() *Config {
	cf := new(Config)
	cf.APIKey = ""
	cf.RealmsList = []string{"eu:fordragon"}
	cf.LocalesList = []string{"en_US", "ru_RU"}
	cf.DownloadDirectory = "data/download"
	cf.TempDirectory =  "data/tmp"
	cf.ResultDirectory =  "data/result"
return cf
}

func (cf *Config) Dump() {
	log.Println("APIKey: ", cf.APIKey)
	log.Println("RealmsList: ", cf.RealmsList)
	log.Println("LocalesList: ", cf.LocalesList)
	log.Println("DownloadDirectory: ", cf.DownloadDirectory)
	log.Println("TempDirectory: ", cf.TempDirectory)
	log.Println("ResultDirectory: ", cf.ResultDirectory)
}

func fixF(name string, defname string, basedir string) string {
	if name == "" {
		name = defname
	}
	if !filepath.IsAbs(name) {
		name = basedir + name
	}
	name, _ = filepath.Abs(name)
	return name
}

func fixD(name string, defname string, basedir string) string {
	name = fixF(name, defname, basedir)
	if name != "" && name[len(name)-1] != SLASH {
		name = name + string(SLASH)
	}
	return name
}

func load(fname string) (*Config, error) {
	dflt := defaultConfig()
	cf := new(Config)
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, cf)
	if err != nil {
		return nil, err
	}
	basedir, err := filepath.Abs(filepath.Dir(fname))
	if err != nil {
		return nil, err
	}
	basedir = basedir + string(SLASH)
	cf.DownloadDirectory = fixD(cf.DownloadDirectory, dflt.DownloadDirectory, basedir)
	cf.TempDirectory = fixD(cf.TempDirectory, dflt.TempDirectory, basedir)
	cf.ResultDirectory = fixD(cf.ResultDirectory, dflt.ResultDirectory, basedir)
	cf.Dump()
	return cf, nil
}

func AppConfig() (*Config, error) {
	dir_base := util.AppDir()
	log.Println("app dir   : ", dir_base)
	r, _ := regexp.Compile("^(.*?)(?:\\.exe|\\.EXE|)$")
	s := r.FindStringSubmatch(util.ExeName())
	cfg_fname :=  s[1] + ".config.json"
	log.Println("config    : ", cfg_fname)
	cf, err := load(cfg_fname)
	if err != nil {
		log.Fatalln("config load error: ", err)
		return nil, err // unreachable
	}
	return cf, nil
}
