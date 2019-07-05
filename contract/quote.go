package contract

import "time"

// Quote is info about a quote
type Quote struct {
	ID        int64
	Text      string
	SpokenBy  []Person
	CreatedOn *time.Time
}

func (q Quote) String() string {
	str := "> *\"" + q.Text + "\"*"
	if q.CreatedOn != nil {
		str += "\n>_(" + q.CreatedOn.Format(time.ANSIC) + ")_"
	}
	for _, p := range q.SpokenBy {
		str += "\n>\t- " + p.Name
	}
	return str
}
