package dao

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/tomlazar/quotes_graph/contract"
)

// PersonDao is a dao that contains operations on people
type PersonDao Dao

// List returns a list of all the users
func (p *PersonDao) List(opts *ListOptions) ([]contract.Person, error) {
	session, err := p.driver.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	records := []contract.Person{}
	_, err = session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		command := withOptions("MATCH (n:Person) RETURN id(n) as ID, n.Name as Name", opts)

		result, err := transaction.Run(command, nil)
		if err != nil {
			return nil, err
		}

		for result.Next() {
			person, err := processPerson(result.Record())
			if err != nil {
				return nil, err
			}

			records = append(records, person)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (p *PersonDao) mergePerson(transaction neo4j.Transaction, person contract.Person) error {
	_, err := transaction.Run(`MERGE (p:Person {Name: $name})`, map[string]interface{}{"name": person.Name})
	return err
}

func processPerson(record neo4j.Record) (contract.Person, error) {
	person, ok := record.Get("Name")
	if !ok {
		return contract.Person{}, errors.New("Could not get person name from the result set")
	}

	id, ok := record.Get("ID")
	if !ok {
		return contract.Person{}, errors.New("Could not get person id from the result set")
	}

	return contract.Person{
		ID:   id.(int64),
		Name: person.(string),
	}, nil
}
