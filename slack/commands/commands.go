package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/contract"
	"github.com/tomlazar/quotes_graph/dao"

	"github.com/getsentry/raven-go"
	"github.com/shomali11/slacker"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = config.NewLogger("commands.commands")
}

// NewBot creates a new bot and initialized all the commands
func NewBot(dao *dao.Dao) *slacker.Slacker {
	bot := slacker.NewClient(viper.GetString("slack.token"))

	bot.Command("list all", ListAllDefinition(dao))
	bot.Command("search <query>", SearchDefinition(dao))
	bot.Command("create <quote> <authors>", CreateDefinition(dao))

	return bot
}

func trim(str string) string {
	return strings.Trim(str, "\"")
}

func reportError(err error, response slacker.ResponseWriter) {
	raven.CaptureError(err, nil)
	logger.Error(err)

	response.ReportError(err)
}

func listQuotes(quotes []contract.Quote, response slacker.ResponseWriter) {
	response.Reply(fmt.Sprintf("Found %v Quotes", len(quotes)))

	for _, item := range quotes {
		response.Reply(item.String())
	}
}

func ListAllDefinition(dao *dao.Dao) *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "List All",
		Example:     "list all",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			quotes, err := dao.QuoteDao.List(nil)

			if err != nil {
				reportError(err, response)
				return
			}

			listQuotes(quotes, response)
		},
	}
}

func SearchDefinition(dao *dao.Dao) *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Search will filter the list from the database and return that list",
		Example:     "search [QUERY]",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			query := trim(request.StringParam("query", ""))

			if query == "" {
				response.ReportError(errors.New("querys must contain some text"))
				return
			}

			paramed := logger.With("query", query, "user", request.Event().Username)

			paramed.Debug("starting query")
			t1 := time.Now()

			quotes, err := dao.QuoteDao.Search(query, nil)

			if err != nil {
				reportError(err, response)
			}

			listQuotes(quotes, response)

			paramed.Debugw("search done", "total time", time.Since(t1))
		},
	}
}

func CreateDefinition(dao *dao.Dao) *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Create adds a new quote to the database. [Authors] is a comma seperated list",
		Example:     "create [QUOTE] [AUTHORS]",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			quote := trim(request.StringParam("quote", ""))
			authors := trim(request.StringParam("authors", ""))
			authorsList := strings.Split(authors, ",")

			if quote == "" {
				response.ReportError(errors.New("quotes must contain some text"))
				return
			}

			if authors == "" {
				response.ReportError(errors.New("authors must contain some text"))
				return
			}

			paramed := logger.With("quote", quote, "authors", authors, "user", request.Event().Username)

			paramed.Debug("starting creation")
			t1 := time.Now()

			quoteObj := contract.Quote{
				CreatedOn: &t1,
				SpokenBy:  []contract.Person{},
				Text:      quote,
			}

			for _, item := range authorsList {
				quoteObj.SpokenBy = append(quoteObj.SpokenBy, contract.Person{Name: item})
			}

			err := dao.QuoteDao.Create(quoteObj)
			if err != nil {
				reportError(err, response)
			}

			response.Reply("New Quote Created!")
			response.Reply(quoteObj.String())

			paramed.Debugw("creation done", "total time", time.Since(t1))
		},
	}
}
