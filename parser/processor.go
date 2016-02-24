package parser

import (
	"time"
	"log"
	"strings"
	"encoding/json"
	"os"
	util "github.com/gourytch/gowowuction/util"
	config "github.com/gourytch/gowowuction/config"
)

type AuctionMeta struct {
	Created  time.Time `json:"created"`
	DeadLine time.Time `json:"deadline"`
	LastSeen time.Time `json:"lastSeen"`
	Updated  time.Time `json:"updated"`
	Raised bool `json:"raised"` // bid change detected
	Bought bool `json:"bought"` // buyout detected
	Expired bool `json:"expired"` // auction definitely not bought
	Moved bool `json:"moved"` // player renamed / moved
	FirstBid int64 `json:"firstBid"`
	LastBid int64 `json:"lastBid"`
}

type WorkEntry struct {
	Entry Auction `json:"entry"`
	Meta  AuctionMeta `json:"meta"`
}

type WorkSetType map[int64]WorkEntry
type IdSetType map[int64]bool

type AuctionProcessorState struct {
	Realm string `json:"realm"`
	LastTime time.Time `json:"lastTime"`
	WorkSet WorkSetType `json:"workSet"`
}

type AuctionProcessor struct {
	StateFName string
	MetaFName string
	AucFName string
	State AuctionProcessorState
	SnapshotTime time.Time
	Started bool
	SeenSet IdSetType
	FileMeta os.File
	FileAuc os.File
	NumCreated int
	NumModified int	
}


const (
	S_VERY_LONG = "VERY_LONG"
	S_LONG = "LONG"
	S_MEDIUM = "MEDIUM"
	S_SHORT = "SHORT"
)

func getTimeInterval(lenStr string, biggest bool) int64 {
	switch {
		case lenStr == S_VERY_LONG:
			if biggest {
				return 2 * 24 * 60 * 60 // 2 days
			} else {
				return 12 * 60 * 60 // 12 hours
			}
		case lenStr == S_LONG:
			if biggest {
				return 12 * 60 * 60 // 12 hours
			} else {
				return 2 * 60 * 60 // 2 hours
			}
		case lenStr == S_MEDIUM:
			if biggest {
				return 2 * 60 * 60 // 2 hours
			} else {
				return 30 * 60 // 30 minutes
			}
		case lenStr == S_SHORT:
			if biggest {
				return 30 * 60 // 30 minutes
			} else {
				return 0
			}
		default:
			log.Fatalf("unknown time interval string <<%s>>", lenStr)
	}
	return 0
}

func calcDeadLine(from time.Time) {
	
}

func (prc *AuctionProcessor) createEntry(auc *Auction) {
	var e WorkEntry
	e.Entry = *auc
	e.Meta.Created = prc.SnapshotTime
	e.Meta.LastSeen = prc.SnapshotTime
	e.Meta.DeadLine = 
	e.Meta.FirstBid = auc.Bid
	e.Meta.LastBid = auc.Bid
	prc.State.WorkSet[e.Entry.Auc] = e
	prc.SeenSet[auc.Auc] = false
	prc.NumCreated++
}

func (prc *AuctionProcessor) applyEntry(auc *Auction) {
	e := prc.State.WorkSet[auc.Auc]
	e.Meta.LastSeen = prc.SnapshotTime
	changed := false
	if auc.Bid != e.Meta.LastBid {
		e.Meta.LastBid = auc.Bid
		e.Entry.Bid = auc.Bid
		changed = true
	}
	if auc.TimeLeft != e.Entry.TimeLeft {
		e.Entry.TimeLeft = auc.TimeLeft
		interval := getTimeInterval(auc.TimeLeft, true) // maximal value
		e.Meta.DeadLine = guessDeadLine(prc.SnapshotTime, prc.SnapshotTime)
		changed = true
	}
	
	prc.State.WorkSet[auc.Auc] = e
	prc.SeenSet[auc.Auc] = false
}

func (prc *AuctionProcessor) closeEntry(id int64) {
	e := prc.State.WorkSet[id]
	delete(prc.State.WorkSet, id)
	e.Meta.Bought = prc.SnapshotTime.Before(e.Meta.DeadLine)
	data_meta, err := json.Marshal(e.Meta)
	if err != nil {
		log.Fatalf("marshall error: %s", err)
	}
	_, err = prc.FileMeta.WriteString(string(data_meta) + "\n")
	if err != nil {
		log.Fatalf("WriteString error: %s", err)
	}
}

func (prc *AuctionProcessor) processAuction(auc *Auction) {
	if _, exists := prc.State.WorkSet[auc.Auc]; exists {
		// modify exists auction
		prc.applyEntry(auc)
	} else {
		prc.createEntry(auc)
	}
}

func (prc *AuctionProcessor) Init(cf *config.Config, realm string) {
	pfx := cf.ResultDirectory + string(config.SLASH) + strings.Replace(realm, ":", "-", 0)
	prc.StateFName = pfx + ".state"
	prc.MetaFName = pfx + ".metadata"
	prc.AucFName = pfx + ".auctions"
	prc.Started = false
}

func (prc *AuctionProcessor) LoadState() {
	if prc.Started {
		log.Fatalln("LoadState inside snapshot session")
	}
	if util.CheckFile(prc.StateFName) {
		log.Printf("AuctionPrrocessor loading state from %s ...", prc.StateFName)
		data, _ := util.Load(prc.StateFName)
		if err := json.Unmarshal(data,  &prc.State); err != nil {
			log.Fatalf("... failed: %s", prc.StateFName, err)
		}
	}
}

func (prc *AuctionProcessor) SaveState() {
	if prc.Started {
		log.Fatalln("SaveState inside snapshot session")
	}
	log.Printf("AuctionPrrocessor storing state to %s ...", prc.StateFName)
	data, err := json.Marshal(&prc.State)
	if err != nil {
		log.Fatalf("... failed: %s", err)
	}
	if strings.HasSuffix(prc.StateFName, ".gz") {
		zdata := util.Zip(data)
		log.Printf("store gzipped (%d%%) data to %s...", 
			len(zdata) * 100 / len(data), prc.StateFName)
		util.Store(prc.StateFName, zdata)
	} else {
		log.Printf("store ungzipped data to %s...", prc.StateFName)
		util.Store(prc.StateFName, data)
	}
}
		
func (prc *AuctionProcessor) StartSnapshot(snaptime time.Time) {
	if prc.Started {
		log.Fatalln("StartSnapshot inside snapshot session")
	}
	prc.Started = true
	prc.SnapshotTime = snaptime
	prc.SeenSet = make(IdSetType)
}

func (prc *AuctionProcessor) AddAuctionEntry(auc *Auction) {
	if !prc.Started {
		log.Fatalln("AddAuctionEntry outside snapshot session")
	}
	prc.processAuction(auc)
}

func (prc *AuctionProcessor) FinishSnapshot() {
	if !prc.Started {
		log.Fatalln("FinishSnapshot outside snapshot session")
	}

	log.Println("check for closed auctions")
	num_open, num_closed, num_changed := 0, 0, 0
	prc.FileAuc, err = os.OpenFile(prc.AucFName, os.O_WRONLY | os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("OpenFile(%s) error: %s", prc.AucFName, err)
	}
	
	prc.FileMeta, err = os.OpenFile(prc.MetaFName, os.O_WRONLY | os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("OpenFile(%s) error: %s", prc.MetaFName, err)
	}
	defer prc.FileMeta.Close()
	for id, _ := range prc.State.WorkSet {
		changed, seen := prc.SeenSet[id]
		if !seen {
			num_closed++
			prc.closeEntry(id)
		} else {
			num_open++
			if changed {
				num_changed++
			}
		}
	}
	log.Printf("%d in open set, %d closed, %d changed", 
	num_open, num_closed, num_changed)
	prc.Started = false
}
