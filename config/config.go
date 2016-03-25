package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
//	"regexp"
	"strings"
	"time"

	util "github.com/gourytch/gowowuction/util"
)

const SLASH = filepath.Separator

type Config struct {
	APIKey            string   `json:"apikey"`
	RealmsList        []string `json:"realms"`
	LocalesList       []string `json:"locales"`
	DownloadDirectory string   `json:"download_dir"`
	TempDirectory     string   `json:"temp_dir"`
	ResultDirectory   string   `json:"result_dir"`
	NameFormat        string   `json:"name_format"`
	TimedNameFormat   string   `json:"timed_name_format"`
}

func defaultConfig() *Config {
	cf := new(Config)
	cf.APIKey = ""
	cf.RealmsList = []string{"eu:fordragon"}
	cf.LocalesList = []string{"en_US", "ru_RU"}
	cf.DownloadDirectory = "data/download"
	cf.TempDirectory = "data/tmp"
	cf.ResultDirectory = "data/result"
	cf.NameFormat = "{realm}-{name}"
	cf.TimedNameFormat = "2006_01-{realm}-{name}" // split by month
	return cf
}

func (cf *Config) Dump() {
	log.Println("APIKey: ", cf.APIKey)
	log.Println("RealmsList: ", cf.RealmsList)
	log.Println("LocalesList: ", cf.LocalesList)
	log.Println("DownloadDirectory: ", cf.DownloadDirectory)
	log.Println("TempDirectory: ", cf.TempDirectory)
	log.Println("ResultDirectory: ", cf.ResultDirectory)
	log.Println("NameFormat:", cf.NameFormat)
	log.Println("TimedNameFormat:", cf.TimedNameFormat)
}

func (cf *Config) GetTimedName(name string, realm string, ts time.Time) string {
	s := ts.Format(cf.TimedNameFormat)
	s = strings.Replace(s, "{realm}", util.Safe_Realm(realm), -1)
	s = strings.Replace(s, "{name}", name, -1)
	return s
}

func (cf *Config) GetName(name string, realm string) string {
	s := strings.Replace(cf.NameFormat, "{realm}", util.Safe_Realm(realm), -1)
	s = strings.Replace(s, "{name}", name, -1)
	return s
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
	if cf.NameFormat == "" {
		cf.NameFormat = dflt.NameFormat
	}
	if cf.TimedNameFormat == "" {
		cf.TimedNameFormat = dflt.TimedNameFormat
	}

	cf.Dump()
	return cf, nil
}

func AppConfig() (*Config, error) {
	cfg_fname := util.AppBaseFileName() + ".config.json"
	log.Println("config    : ", cfg_fname)
	cf, err := load(cfg_fname)
	if err != nil {
		log.Fatalln("config load error: ", err)
		return nil, err // unreachable
	}
	return cf, nil
}
