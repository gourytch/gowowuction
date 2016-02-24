package parser

import (
	"encoding/json"
	"log"
)

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
