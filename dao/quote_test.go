package dao

import (
	"testing"

	_ "github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/contract"
)

func TestQuoteDao_Create(t *testing.T) {
	dao, err := NewDao()
	if err != nil {
		t.Fatal(err)
	}

	err = dao.QuoteDao.Create(contract.Quote{
		Text: "Demo2",
		SpokenBy: []contract.Person{
			contract.Person{
				Name: "Test",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
