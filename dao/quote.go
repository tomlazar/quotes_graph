package dao

import (
	"errors"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/tomlazar/quotes_graph/contract"
)

// QuoteDao is a dao that contains operations on quotes
type QuoteDao Dao

var timefmt = time.ANSIC

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

	createdOn, ok := reader.Get("CreatedOn")
	if !ok {
		return nil, errors.New("Could not get created on from result set")
	}
	if createdOn != nil {
		createdOnString := createdOn.(string)
		createdOnTime, err := time.Parse(timefmt, createdOnString)
		if err != nil {
			return nil, errors.New("Could not parse string to time: " + createdOn.(string))
		}
		record.CreatedOn = &createdOnTime
	} else {
		record.CreatedOn = nil
	}

	record.SpokenBy = []contract.Person{}
	personResult, err := transaction.Run(
		`MATCH (p:Person)<-[:SPOKEN_BY]-(:Quote {Text: $text})
		 RETURN id(p) as ID, p.Name as Name`, map[string]interface{}{"text": record.Text})
	if err != nil {
		return nil, err
	}

	for personResult.Next() {
		person, ok := personResult.Record().Get("Name")
		if !ok {
			return nil, errors.New("Could not get person name from the result set")
		}

		id, ok := personResult.Record().Get("ID")
		if !ok {
			return nil, errors.New("Could not get person id from the result set")
		}

		record.SpokenBy = append(record.SpokenBy, contract.Person{
			ID:   id.(int64),
			Name: person.(string),
		})
	}

	return &record, nil
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
		command := withOptions("MATCH (n:Quote) RETURN ID(n) as ID, n.Text as Text, n.CreatedOn as CreatedOn ORDER BY ID(n) ", opts)

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
		command := withOptions("MATCH (n:Quote) WHERE n.Text =~ $search RETURN ID(n) as ID, n.Text as Text, n.CreatedOn as CreatedOn ORDER BY ID(n) ", opts)

		result, err := transaction.Run(command, map[string]interface{}{"search": "(?i)" + s})

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

// Create will create a new quote in the database
func (q *QuoteDao) Create(quote contract.Quote) error {
	session, err := q.driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		for _, person := range quote.SpokenBy {
			err = q.PersonDao.mergePerson(transaction, person)
			if err != nil {
				return nil, rollbackWithErr(transaction, err)
			}
		}

		res, err := transaction.Run(
			`CREATE (q:Quote {Text: $text, CreatedOn: $createdOn}) RETURN ID(q) as id`,
			map[string]interface{}{
				"text":      quote.Text,
				"createdOn": quote.CreatedOn.Format(timefmt),
			},
		)

		if !res.Next() || err != nil {
			return nil, rollbackWithErr(transaction, fmt.Errorf("Could not create node: %v", err))
		}

		id := res.Record().GetByIndex(0)

		for _, person := range quote.SpokenBy {
			_, err = transaction.Run(`
			MATCH (q:Quote), (p:Person)
			WHERE ID(q) = $quoteId
			AND p.Name = $name
			CREATE (q)-[:SPOKEN_BY]->(p)
		`, map[string]interface{}{
				"quoteId": id,
				"name":    person.Name,
			})

			if err != nil {
				return nil, rollbackWithErr(transaction, err)
			}
		}

		return nil, nil
	})

	return err
}

func rollbackWithErr(t neo4j.Transaction, e error) error {
	err2 := t.Rollback()

	if err2 != nil {
		return err2
	}

	return e
}
