package parser

import (
	"encoding/json"
	"log"
)

type TimeLeft int

const (
	VERY_LONG TimeLeft = iota
	LONG
	MEDIUM
	SHORT
)

type Bonus struct {
	BonusListId int32 `json:"bonusListId"`
}

type BonusList []Bonus

type Modifier struct {
	Type  int32 `json:"type"`
	Value int32 `json:"value"`
}

type ModList []Modifier

/**/
type BaseAuction struct {
	Auc        int64  `json:"auc"`
	Item       int64  `json:"item"`
	Owner      string `json:"owner"`
	OwnerRealm string `json:"ownerRealm"`
	Bid        int64  `json:"bid"`
	Buyout     int64  `json:"buyout"`
	Quantity   int32  `json:"quantity"`
	TimeLeft   string `json:"timeLeft"` // VERY_LONG | LONG | MEDIUM | SHORT
	Rand       int64  `json:"rand"`
	Seed       int64  `json:"seed"`
	Context    int64  `json:"context"`
}

type ModsPart struct {
	Modifiers ModList `json:"modifiers"`
}

type BonusPart struct {
	BonusLists BonusList `json:"bonusLists"`
}

type PetPart struct {
	PetSpeciesId int `json:"petSpeciesId"`
	PetBreedId   int `json:"petBreedId"`
	PetLevel     int `json:"petLevel"`
	PetQualityId int `json:"petQualityId"`
}

type AuctionWithBonus struct {
	BaseAuction
	BonusPart
}

type AuctionWithMods struct {
	BaseAuction
	BonusPart
	ModsPart
}

type PetAuction struct {
	BaseAuction
	ModsPart
	PetPart
}

type Auction struct {
	BaseAuction
	ModsPart
	BonusPart
	PetPart
}

/**/

/**
type Auction struct {
	// base part
	Auc        int64  `json:"auc"`
	Item       int64  `json:"item"`
	Owner      string `json:"owner"`
	OwnerRealm string `json:"ownerRealm"`
	Bid        int64  `json:"bid"`
	Buyout     int64  `json:"buyout"`
	Quantity   int32  `json:"quantity"`
	TimeLeft   string `json:"timeLeft"` // VERY_LONG | LONG | MEDIUM | SHORT
	Rand       int64  `json:"rand"`
	Seed       int64  `json:"seed"`
	Context    int64  `json:"context"`
	// mods part
	Modifiers ModList `json:"modifiers"`
	// bonus part
	BonusLists BonusList `json:"bonusLists"`
	// Pet Part
	PetSpeciesId int `json:"petSpeciesId"`
	PetBreedId   int `json:"petBreedId"`
	PetLevel     int `json:"petLevel"`
	PetQualityId int `json:"petQualityId"`
}
**/

type Realm struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type RawAuctionData map[string]interface{}

type SnapshotData struct {
	Realms   []Realm   `json:"realms"`
	Auctions []Auction `json:"auctions"`
}

/*
func ParseAuction(data *RawAuctionData) (r *Auction) {
	r = new(Auction)
	r.Auc = (*data)["auc"].(uint64)
	r.Item = (*data)["item"].(uint64)
	r.Owner = (*data)["owner"].(string)
	r.Realm = (*data)["ownerRealm"].(string)
	r.Bid = (*data)["bid"].(uint64)
	r.Buyout = (*data)["buyout"].(uint64)
	r.Quantity = (*data)["quantity"].(uint32)
	r.TimeLeft = (*data)["timeLeft"].(string)
	r.Rand = (*data)["rand"].(int64)
	r.Seed = (*data)["seed"].(int64)
	r.Context = (*data)["context"].(int64)
	return
}
*/

func ParseSnapshot(data []byte) (snapshot *SnapshotData) {
	snapshot = new(SnapshotData)
	if err := json.Unmarshal(data, snapshot); err != nil {
		log.Fatalf("... json failed: %s", err)
	}
	return
}

func MakeBaseAuction(auc *Auction) (bse *BaseAuction) {
	bse = new(BaseAuction)
	*bse = auc.BaseAuction
	return
}

func MakeAuctionWithBonus(auc *Auction) (bns *AuctionWithMods) {
	bns = new(AuctionWithMods)
	bns.BaseAuction = auc.BaseAuction
	bns.BonusPart = auc.BonusPart
	return
}

func MakeAuctionWithMods(auc *Auction) (mod *AuctionWithMods) {
	mod = new(AuctionWithMods)
	mod.BaseAuction = auc.BaseAuction
	mod.ModsPart = auc.ModsPart
	mod.BonusPart = auc.BonusPart
	return
}

func MakePetAuction(auc *Auction) (pet *PetAuction) {
	pet = new(PetAuction)
	pet.BaseAuction = auc.BaseAuction
	pet.ModsPart = auc.ModsPart
	pet.PetPart = auc.PetPart
	return
}

func PackAuctionData(auc *Auction) (blob []byte) {
	switch {
	case auc.PetSpeciesId != 0:
		blob, _ = json.Marshal(MakePetAuction(auc))
	case auc.Modifiers != nil:
		blob, _ = json.Marshal(MakeAuctionWithMods(auc))
	case auc.BonusLists != nil:
		blob, _ = json.Marshal(MakeAuctionWithBonus(auc))
	default:
		blob, _ = json.Marshal(MakeBaseAuction(auc))
	}
	return
}
