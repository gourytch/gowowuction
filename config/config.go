package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Config struct {
	APIKey                         string   `json:"apikey"`
	RealmsList                     []string `json:"realms"`
	LocalesList                    []string `json:"locales"`
	DownloadDirectory              string   `json:"download_dir"`
	TempDirectory                  string   `json:"temp_dir"`
	ResultDirectory                string   `json:"data_dir"`
	OpenedAuctionsFilename         string   `json:"opened_name"`
	ClosedAuctionsFilename         string   `json:"closed_name"`
	ClosedAuctionsMetadataFilename string   `json:"meta_name"`
}

func Default() *Config {
	cf := new(Config)
	cf.APIKey = ""
	cf.RealmsList = []string{"eu:fordragon"}
	cf.LocalesList = []string{"en_US", "ru_RU"}
	cf.DownloadDirectory = "./data/download"
	cf.TempDirectory = "./data/tmp"
	cf.ResultDirectory = "./data/results"
	cf.OpenedAuctionsFilename = "open.data"
	cf.ClosedAuctionsFilename = "closed.data"
	cf.ClosedAuctionsMetadataFilename = "closed-metadata.data"
	return cf
}

func (cf *Config) Dump() {
	log.Println("APIKey: ", cf.APIKey)
	log.Println("RealmsList: ", cf.RealmsList)
	log.Println("LocalesList: ", cf.LocalesList)
	log.Println("DownloadDirectory: ", cf.DownloadDirectory)
	log.Println("TempDirectory: ", cf.TempDirectory)
	log.Println("ResultDirectory: ", cf.ResultDirectory)
	log.Println("OpenedAuctionsFilename: ", cf.OpenedAuctionsFilename)
	log.Println("ClosedAuctionsFilename: ", cf.ClosedAuctionsFilename)
	log.Println("ClosedAuctionsMetadataFilename: ", cf.ClosedAuctionsMetadataFilename)
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
	if name != "" && name[len(name)-1] != filepath.Separator {
		name = name + string(filepath.Separator)
	}
	return name
}

func Load(fname string) (*Config, error) {
	dflt := Default()
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
	basedir = basedir + string(filepath.Separator)

	cf.DownloadDirectory = fixD(cf.DownloadDirectory, dflt.DownloadDirectory, basedir)
	cf.TempDirectory = fixD(cf.TempDirectory, dflt.TempDirectory, basedir)
	cf.ResultDirectory = fixD(cf.ResultDirectory, dflt.ResultDirectory, basedir)
	cf.OpenedAuctionsFilename = fixF(cf.OpenedAuctionsFilename, dflt.OpenedAuctionsFilename, basedir)
	cf.ClosedAuctionsFilename = fixF(cf.ClosedAuctionsFilename, dflt.ClosedAuctionsFilename, basedir)
	cf.ClosedAuctionsMetadataFilename = fixF(cf.ClosedAuctionsMetadataFilename, dflt.ClosedAuctionsMetadataFilename, basedir)

	cf.Dump()
	return cf, nil
}
