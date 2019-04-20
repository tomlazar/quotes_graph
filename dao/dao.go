package dao

import (
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/spf13/viper"
)

// Dao contains information about how to get data from the database
type Dao struct {
	driver neo4j.Driver

	QuoteDao *QuoteDao
}

func init() {
	viper.SetDefault("neo.uri", nil)
	viper.SetDefault("neo.auth.username", nil)
	viper.SetDefault("neo.auth.password", nil)
}

// NewDao will create a new dao with the driver initialized
func NewDao() (*Dao, error) {
	uri := viper.Get("neo.uri")
	if uri == nil {
		return nil, fmt.Errorf("neo4j uri is not defined")
	}

	username := viper.Get("neo.auth.username")
	if username == nil {
		return nil, fmt.Errorf("neo4j username is not defined")
	}

	password := viper.Get("neo.auth.password")
	if password == nil {
		return nil, fmt.Errorf("neo4j password is not defined")
	}

	logrus.WithField("uri", uri).Debugln("Starting new Dao")
	driver, err := neo4j.NewDriver(uri.(string), neo4j.BasicAuth(username.(string), password.(string), ""))
	if err != nil {
		return nil, err
	}

	dao := &Dao{
		driver: driver,
	}

	dao.QuoteDao = (*QuoteDao)(dao)

	return dao, nil
}
