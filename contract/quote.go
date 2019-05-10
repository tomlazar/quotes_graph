package contract

import "time"

// Quote is info about a quote
type Quote struct {
	ID        int64
	Text      string
	SpokenBy  []Person
	CreatedOn *time.Time
}
