package parser

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	config "github.com/gourytch/gowowuction/config"
	util "github.com/gourytch/gowowuction/util"
)

type AuctionMeta struct {
	Created  time.Time `json:"created"`
	DeadLine time.Time `json:"deadline"`
	LastSeen time.Time `json:"lastSeen"`
	Updated  time.Time `json:"updated"`
	Raised   bool      `json:"raised"`  // bid change detected
	Bought   bool      `json:"bought"`  // buyout detected
	Expired  bool      `json:"expired"` // auction definitely not bought
	Moved    bool      `json:"moved"`   // player renamed / moved
	FirstBid int64     `json:"firstBid"`
	LastBid  int64     `json:"lastBid"`
}

type WorkEntry struct {
	Entry Auction     `json:"entry"`
	Meta  AuctionMeta `json:"meta"`
}

type WorkSetType map[int64]WorkEntry
type WorkListType []WorkEntry
type IdSetType map[int64]bool

type AuctionProcessorState struct {
	Realm    string       `json:"realm"`
	LastTime time.Time    `json:"lastTime"`
	WorkSet  WorkSetType  `json:"-"`
	WorkList WorkListType `json:"worklist"`
}

type AuctionProcessor struct {
	cf           *config.Config
	StateFName   string
	Realm        string
	State        AuctionProcessorState
	SnapshotTime time.Time
	Started      bool
	SeenSet      IdSetType
	FileMeta     *os.File
	FileAuc      *os.File
	NumCreated   int
	NumModified  int
	NumBids      int
	NumMoves     int
	NumAdjusts   int
}

const (
	S_VERY_LONG = "VERY_LONG"
	S_LONG      = "LONG"
	S_MEDIUM    = "MEDIUM"
	S_SHORT     = "SHORT"
)

func get_expiration_interval(exp string) (min, max time.Duration) {
	switch {
	case exp == S_SHORT: // "SHORT" -> 0 .. 30m
		return 0, 30 * time.Minute
	case exp == S_MEDIUM: // "MEDIUM" -> 30m .. 2h
		return 30 * time.Minute, 2 * time.Hour
	case exp == S_LONG: // "LONG" -> 2h .. 12h
		return 2 * time.Hour, 12 * time.Hour
	case exp == S_VERY_LONG: // "VERY_LONG" -> 12h .. 2d
		return 12 * time.Hour, 48 * time.Hour
	default:
		log.Fatalf("unknown expiration time string <<%s>>", exp)
	}
	return
}

func random_duration(d time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(d)))
}

func random_datetime(a, b time.Time) time.Time {
	if b.Before(a) {
		a, b = b, a
	}
	return a.Add(random_duration(b.Sub(a)))
}

func guess_expiration(t time.Time, exp string) (min, max time.Time) {
	dmin, dmax := get_expiration_interval(exp)
	return t.Add(dmin), t.Add(dmax)
}

func (prc *AuctionProcessor) createEntry(auc *Auction) {
	id := auc.Auc
	var e WorkEntry
	e.Entry = *auc
	e.Meta.Created = prc.SnapshotTime
	e.Meta.LastSeen = prc.SnapshotTime
	dl_min, _ := guess_expiration(prc.SnapshotTime, e.Entry.TimeLeft)
	var zeroTime time.Time
	if prc.State.LastTime == zeroTime { // zero value
		e.Meta.DeadLine = dl_min
	} else { // assigned
		_, dl_max2 := guess_expiration(prc.State.LastTime, e.Entry.TimeLeft)
		if dl_min.Before(dl_max2) {
			e.Meta.DeadLine = dl_min
		} else {
			e.Meta.DeadLine = dl_max2
		}
	}
	e.Meta.FirstBid = auc.Bid
	e.Meta.LastBid = auc.Bid
	prc.State.WorkSet[id] = e
	prc.SeenSet[id] = false
	prc.NumCreated++
}

func (prc *AuctionProcessor) applyEntry(auc *Auction) {
	id := auc.Auc
	e := prc.State.WorkSet[id]
	e.Meta.LastSeen = prc.SnapshotTime
	changed := false
	if auc.Bid != e.Meta.LastBid {
		e.Meta.LastBid = auc.Bid
		e.Entry.Bid = auc.Bid
		e.Meta.Raised = true
		prc.NumBids++
		changed = true
	}
	if auc.TimeLeft != e.Entry.TimeLeft {
		e.Entry.TimeLeft = auc.TimeLeft
		_, e.Meta.DeadLine = guess_expiration(prc.SnapshotTime, e.Entry.TimeLeft)
		prc.NumAdjusts++
		changed = true
	}
	if auc.Owner != e.Entry.Owner || auc.OwnerRealm != e.Entry.OwnerRealm {
		e.Entry.Owner = auc.Owner
		e.Entry.OwnerRealm = auc.OwnerRealm
		e.Meta.Moved = true
		prc.NumMoves++
		changed = true
	}

	prc.State.WorkSet[id] = e
	prc.SeenSet[id] = changed
	if changed {
		prc.NumModified++
	}
}

func (prc *AuctionProcessor) closeEntry(id int64) {
	e := prc.State.WorkSet[id]
	delete(prc.State.WorkSet, id)
	e.Meta.Bought = prc.SnapshotTime.Before(e.Meta.DeadLine)
	data_auc, err := json.Marshal(e.Entry)
	data_meta, err := json.Marshal(e.Meta)
	if err != nil {
		log.Panicf("marshall error: %s", err)
	}
	_, err = prc.FileAuc.WriteString(string(data_auc) + "\n")
	_, err = prc.FileMeta.WriteString(string(data_meta) + "\n")
	if err != nil {
		log.Panicf("WriteString error: %s", err)
	}
}

func (prc *AuctionProcessor) processAuction(auc *Auction) {
	id := auc.Auc
	if _, exists := prc.State.WorkSet[id]; exists {
		// modify exists auction
		prc.applyEntry(auc)
	} else {
		prc.createEntry(auc)
	}
}

func (prc *AuctionProcessor) Init(cf *config.Config, realm string) {
	prc.cf = cf
	prc.Realm = realm
	prc.StateFName = cf.ResultDirectory + cf.GetName("state", prc.Realm) + ".gz"
	prc.State.WorkSet = make(WorkSetType)
	prc.State.WorkList = nil
	prc.SnapshotTime = time.Time{}
	prc.Started = false
	prc.SeenSet = make(IdSetType)
	prc.FileMeta = nil
	prc.FileAuc = nil
	prc.NumCreated = 0
	prc.NumModified = 0
	prc.NumBids = 0
	prc.NumMoves = 0
	prc.NumAdjusts = 0
}

func (prc *AuctionProcessor) LoadState() {
	if prc.Started {
		log.Panic("LoadState inside snapshot session")
	}
	if util.CheckFile(prc.StateFName) {
		log.Printf("AuctionProcessor loading state from %s ...", prc.StateFName)
		data, _ := util.Load(prc.StateFName)
		if err := json.Unmarshal(data, &prc.State); err != nil {
			log.Panicf("... failed: %s", prc.StateFName, err)
		}
		log.Printf("... loaded with %d list enties", len(prc.State.WorkList))
		prc.State.WorkSet = make(WorkSetType)
		for _, e := range prc.State.WorkList {
			prc.State.WorkSet[e.Entry.Auc] = e
		}
	} else {
		log.Printf("AuctionProcessor has no state named %s ...", prc.StateFName)
	}
}

func (prc *AuctionProcessor) SaveState() {
	if prc.Started {
		log.Panic("SaveState inside snapshot session")
	}
	log.Printf("AuctionProcessor storing state to %s ...", prc.StateFName)
	log.Printf("... prepare list with %d enties", len(prc.State.WorkSet))
	prc.State.WorkList = WorkListType{}
	for _, e := range prc.State.WorkSet {
		prc.State.WorkList = append(prc.State.WorkList, e)
	}
	data, err := json.Marshal(&prc.State)
	if err != nil {
		log.Fatalf("... failed: %s", err)
	}
	if strings.HasSuffix(prc.StateFName, ".gz") {
		zdata := util.Zip(data)
		log.Printf("store gzipped (%d%%) data to %s...",
			len(zdata)*100/len(data), prc.StateFName)
		util.Store(prc.StateFName, zdata)
	} else {
		log.Printf("store ungzipped data to %s...", prc.StateFName)
		util.Store(prc.StateFName, data)
	}
}

func (prc *AuctionProcessor) SnapshotNeeded(snaptime time.Time) bool {
	return prc.State.LastTime.Before(snaptime)
}

func (prc *AuctionProcessor) StartSnapshot(snaptime time.Time) {
	if prc.Started {
		log.Panic("StartSnapshot inside snapshot session")
	}
	prc.Started = true
	prc.SnapshotTime = snaptime
	prc.SeenSet = make(IdSetType)
	prc.NumCreated = 0
	prc.NumModified = 0
	prc.NumBids = 0
	prc.NumMoves = 0
	prc.NumAdjusts = 0
	// log.Printf("start snapshot at %s with %d entries in workset",
	//	util.TSStr(prc.SnapshotTime), len(prc.State.WorkSet))
}

func (prc *AuctionProcessor) AddAuctionEntry(auc *Auction) {
	if !prc.Started {
		log.Panic("AddAuctionEntry outside snapshot session")
	}
	prc.processAuction(auc)
}

func OpenOrCreateFile(fname string) *os.File {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		f, err = os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Panicf("OpenFile(%s) error: %s", fname, err)
		}
	}
	return f
}

func (prc *AuctionProcessor) FinishSnapshot() {
	if !prc.Started {
		log.Panic("FinishSnapshot outside snapshot session")
	}

	// log.Println("check for closed auctions")
	num_open, num_closed := 0, 0
	auc_fname := prc.cf.ResultDirectory + prc.cf.GetTimedName("auctions", prc.Realm, prc.SnapshotTime)
	meta_fname := prc.cf.ResultDirectory + prc.cf.GetTimedName("metadata", prc.Realm, prc.SnapshotTime)

	prc.FileAuc = OpenOrCreateFile(auc_fname)
	defer prc.FileAuc.Close()

	prc.FileMeta = OpenOrCreateFile(meta_fname)
	defer prc.FileMeta.Close()

	for id, _ := range prc.State.WorkSet {
		_, seen := prc.SeenSet[id]
		if !seen {
			num_closed++
			prc.closeEntry(id)
		} else {
			num_open++
		}
	}

	log.Printf("%s: %d entries: active %d, created %d, changed %d [bids:%d, adj:%d, moves:%d], closed %d",
		util.TSStr(prc.SnapshotTime),
		len(prc.State.WorkSet), num_open,
		prc.NumCreated, prc.NumModified,
		prc.NumBids, prc.NumAdjusts, prc.NumMoves,
		num_closed)
	prc.State.LastTime = prc.SnapshotTime
	//log.Printf("last time sets to %s", util.TSStr(prc.State.LastTime))
	prc.Started = false
}
