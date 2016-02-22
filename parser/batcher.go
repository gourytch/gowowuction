package parser

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"

	util "github.com/gourytch/gowowuction/util"
)

type ByBasename []string

func (a ByBasename) Len() int           { return len(a) }
func (a ByBasename) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBasename) Less(i, j int) bool { return filepath.Base(a[i]) < filepath.Base(a[j]) }

const TRIM_COUNT = 10000000

func ProcessSnapshot(ss *SnapshotData) {
	log.Printf("snapshot for %d auctions in %d realms",
		len(ss.Auctions), len(ss.Realms))
	log.Printf("  realms:")
	for _, realm := range ss.Realms {
		log.Printf("  name=%s, slug=%s", realm.Name, realm.Slug)
	}
	count := len(ss.Auctions)
	if TRIM_COUNT < count {
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

func BatchParse(fnames []string) {
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
	sort.Sort(ByBasename(goodfnames))
	for _, fname := range fnames {
		log.Println(fname)
		data, err := util.Load(fname)
		if err != nil {
			log.Fatalln("load error:", err)
		}
		ProcessSnapshot(ParseSnapshot(data))
		break
	}
}
