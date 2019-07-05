package dao

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/tomlazar/quotes_graph/config"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/spf13/viper"
)

// Dao contains information about how to get data from the database
type Dao struct {
	driver neo4j.Driver

	QuoteDao  *QuoteDao
	PersonDao *PersonDao
}

var logger *zap.SugaredLogger

func init() {
	logger = config.NewLogger("dao")

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

	logger.Debugw("starting a new dao",
		"uri", uri,
	)

	driver, err := neo4j.NewDriver(uri.(string), neo4j.BasicAuth(username.(string), password.(string), ""))
	if err != nil {
		return nil, err
	}

	dao := &Dao{
		driver: driver,
	}

	dao.QuoteDao = (*QuoteDao)(dao)
	dao.PersonDao = (*PersonDao)(dao)

	return dao, nil
}

// ListOptions are options to refine the results of a list command
type ListOptions struct {
	Skip  int
	Limit int
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
