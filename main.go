package main

import (
	"log"
	"os"
	"path/filepath"

	"./config"
	"./fetcher"
	"./util"
)

func main() {
	log.Println("start")
	dir_base := util.AppDir()
	log.Println("app dir   : ", dir_base)
	var cfg_fname string
	if len(os.Args) > 1 {
		cfg_fname = os.Args[1]
	} else {
		cfg_fname = dir_base + string(filepath.Separator) + "config.json"
	}
	log.Println("config    : ", cfg_fname)
	cf, err := config.Load(cfg_fname)
	if err != nil {
		log.Fatalln("config load error: ", err)
	}

	cf.Dump()

	util.CheckDir(cf.DownloadDirectory)
	util.CheckDir(cf.ResultDirectory)

	s := new(fetcher.Session)
	s.Config = cf
	for _, realm := range cf.RealmsList {
		for _, locale := range cf.LocalesList {
			file_url, file_ts := s.Fetch_FileURL(realm, locale)
			log.Printf("FILE URL: %s", file_url)
			log.Printf("FILE PIT: %s / %s", file_ts, util.TSStr(file_ts.UTC()))
			fname := util.Make_FName(realm, file_ts)
			json_fname := cf.DownloadDirectory + fname
			if !util.CheckFile(json_fname) {
				log.Println("downloading from ", file_url)
				data := s.Get(file_url)
				log.Println("... got ", len(data), " octets")
				zdata := util.Zip(data)
				log.Println("... zipped to ", len(zdata), " octets (",
					len(zdata)*100/len(data), "%)")
				util.Store(json_fname, zdata)
				log.Println("stored to ", json_fname)
			} else {
				log.Println("... already downloaded")
			}
		}
	}
	log.Println("finish")
}
