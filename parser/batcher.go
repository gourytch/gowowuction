package parser

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	config "github.com/gourytch/gowowuction/config"
	util "github.com/gourytch/gowowuction/util"
)

const TRIM_COUNT = 0

func ProcessSnapshot(ss *SnapshotData) {
	log.Printf("snapshot for %d auctions in %d realms",
		len(ss.Auctions), len(ss.Realms))
	log.Printf("  realms:")
	for _, realm := range ss.Realms {
		log.Printf("  name=%s, slug=%s", realm.Name, realm.Slug)
	}
	count := len(ss.Auctions)
	if TRIM_COUNT > 0 && TRIM_COUNT < count {
		count = TRIM_COUNT
	}
	log.Printf("  auctions: %d", count)
	for _, auc := range ss.Auctions {
		//log.Printf("raw=%#v", auc)
		blob := PackAuctionData(&auc)
		fmt.Println(string(blob))
		count--
		if count <= 0 {
			break
		}
	}
}

func ParseDir(cf *config.Config, realm string, safe bool) {
	mask := cf.DownloadDirectory +
		strings.Replace(realm, ":", "-", -1) + "-*.json.gz"
	log.Printf("scan by mask %s ...", mask)
	fnames, err := filepath.Glob(mask)
	if err != nil {
		log.Fatalln("glob failed:", err)
	}
	log.Printf("... %d entries collected", len(fnames))

	var goodfnames []string

	for _, fname := range fnames {
		// realm, ts, good := util.Parse_FName(fname)
		_, _, good := util.Parse_FName(fname)
		if good {
			// log.Printf("fname %s -> %s, %v", fname, realm, ts)
			goodfnames = append(goodfnames, fname)
		} else {
			// log.Printf("skip fname %s", fname)
		}
	}
	sort.Sort(util.ByBasename(goodfnames))
	prc := new(AuctionProcessor)
	prc.Init(cf, realm)
	prc.LoadState()
	badfiles := make(map[string]string)

	for _, fname := range fnames {
		//log.Println(fname)
		f_realm, f_time, ok := util.Parse_FName(fname)
		if !ok {
			log.Fatalf("not parsed correctly: %s", fname)
			continue
		}
		if f_realm != realm {
			log.Fatalf("not my realm (%s != %s)", f_realm, realm)
			continue
		}
		if !prc.SnapshotNeeded(f_time) {
			log.Printf("snapshot not needed: %s", util.TSStr(f_time))
			continue
		}
		data, err := util.Load(fname)
		if err != nil {
			//log.Fatalf("load error: %s", err)
			log.Printf("%s LOAD ERROR: %s", fname, err)
			badfiles[fname] = fmt.Sprint(err)
			continue
		}
		ss, err := ParseSnapshot(data)
		if err != nil {
			//log.Fatalf("load error: %s", err)
			log.Printf("%s PARSE ERROR: %s", fname, err)
			badfiles[fname] = fmt.Sprint(err)
			continue
		}

		prc.StartSnapshot(f_time)
		for _, auc := range ss.Auctions {
			prc.AddAuctionEntry(&auc)
		}
		prc.FinishSnapshot()
		if safe {
			prc.SaveState()
		}
	}
	if !safe {
		prc.SaveState()
	}
	if len(badfiles) == 0 {
		log.Printf("all files loaded without errors")
	} else {
		log.Printf("%d files with errors", len(badfiles))
		for fname, err := range badfiles {
			log.Printf("%s: %s", fname, err)
		}
	}
}
