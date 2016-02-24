package parser

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
