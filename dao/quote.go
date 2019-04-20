package dao

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/tomlazar/quotes_graph/contract"
)

// QuoteDao is a dao that contains operations on quotes
type QuoteDao Dao

// ListOptions are options to refine the results of a list command
type ListOptions struct {
	Skip  int
	Limit int
}

func processQuote(transaction neo4j.Transaction, result neo4j.Result) (*contract.Quote, error) {
	record := contract.Quote{}
	reader := result.Record()

	id, ok := reader.Get("ID")
	if !ok {
		return nil, errors.New("Could not get ID from result set")
	}
	record.ID = id.(int64)

	text, ok := reader.Get("Text")
	if !ok {
		return nil, errors.New("Could not get Text from result set")
	}
	record.Text = text.(string)

	record.SpokenBy = []string{}
	personResult, err := transaction.Run(
		`MATCH (p:Person)<-[:SPOKEN_BY]-(:Quote {Text: $text})
		 RETURN p.Name as Name`, map[string]interface{}{"text": record.Text})
	if err != nil {
		return nil, err
	}

	for personResult.Next() {
		person, ok := personResult.Record().Get("Name")
		if !ok {
			return nil, errors.New("Could not get person name from the result set")
		}
		record.SpokenBy = append(record.SpokenBy, person.(string))
	}

	return &record, nil
}

func withOptions(s string, opts *ListOptions) string {
	if opts != nil {
		if opts.Skip > 0 {
			s += fmt.Sprintf("SKIP %v ", opts.Skip)
		}

		if opts.Limit > 0 {
			s += fmt.Sprintf("LIMIT %v ", opts.Limit)
		}
	}
	return (string)(s)
}

// List will return a list of quotes
func (q *QuoteDao) List(opts *ListOptions) ([]contract.Quote, error) {
	session, err := q.driver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	records := []contract.Quote{}
	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		command := withOptions("MATCH (n:Quote) RETURN ID(n) as ID, n.Text as Text ORDER BY ID(n) ", opts)

		result, err := transaction.Run(command, nil)
		if err != nil {
			return nil, err
		}

		for result.Next() {
			record, err := processQuote(transaction, result)
			if err != nil {
				return nil, err
			}

			records = append(records, *record)
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}

// Search will only search for string that are a match to the query string
func (q *QuoteDao) Search(s string, opts *ListOptions) ([]contract.Quote, error) {
	session, err := q.driver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	records := []contract.Quote{}
	session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		command := withOptions("MATCH (n:Quote) WHERE n.Text =~ $search RETURN ID(n) as ID, n.Text as Text ORDER BY ID(n) ", opts)

		result, err := transaction.Run(command, map[string]interface{}{"search": "(?i)" + s})

		sum, err := result.Summary()

		logrus.Debugln(sum.Statement())

		if err != nil {
			return nil, err
		}

		for result.Next() {
			record, err := processQuote(transaction, result)
			if err != nil {
				return nil, err
			}

			records = append(records, *record)
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}
