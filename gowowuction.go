package main

import (
	"log"

	config "github.com/gourytch/gowowuction/config"
	fetcher "github.com/gourytch/gowowuction/fetcher"
	parser "github.com/gourytch/gowowuction/parser"
	util "github.com/gourytch/gowowuction/util"
)

func DoFetch(cf *config.Config) {
	log.Println("=== FETCH BEGIN ===")
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
				log.Printf("downloading from %s ...", file_url)
				data := s.Get(file_url)
				log.Printf("... got %d octets", len(data))
				zdata := util.Zip(data)
				log.Printf("... zipped to %d octets (%d%%)",
					len(zdata), len(zdata)*100/len(data))
				util.Store(json_fname, zdata)
				log.Printf("stored to %s .", json_fname)
			} else {
				log.Println("... already downloaded")
			}
		}
	}
	log.Println("=== FETCH END ===")
}

func DoParse(cf *config.Config) {
	log.Println("=== PARSE BEGIN ===")
	for _, realm := range cf.RealmsList {
		parser.ParseDir(cf, realm, false)
	}
	log.Println("=== PARSE END ===")
}

func main() {
	log.Println("start")
	cf, err := config.AppConfig()
	if err != nil {
		log.Fatalln("config load error: ", err)
	}

	cf.Dump()

	util.CheckDir(cf.DownloadDirectory)
	util.CheckDir(cf.ResultDirectory)

	//DoFetch(cf)
	DoParse(cf)

	log.Println("done")
}
