package parser

import (
	"time"
)

type AuctionMetadata struct {
	Id       uint64
	Created  time.Time
	Updated  time.Time
	DeadLine time.Time
}

type Processor struct {
}
