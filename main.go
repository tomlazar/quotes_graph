package main

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/getsentry/raven-go"

	"github.com/tomlazar/quotes_graph/config"
	_ "github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/contract"
	"github.com/tomlazar/quotes_graph/dao"

	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

func main() {
	token := viper.GetString("slack.token")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	dao, err := dao.NewDao()
	if err != nil {
		config.Logger.Fatal(err)
	}

	displayErr := func(err error, channel string) {
		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Could not execute: %v", err), channel))
	}

	reportErr := func(err error, channel string) {
		raven.CaptureError(err, nil)
		config.Logger.Error(err)
		displayErr(err, channel)
	}

	formatQuote := func(q contract.Quote) string {
		str := "> *\"" + q.Text + "\"*"
		if q.CreatedOn != nil {
			str += "\n>_(" + q.CreatedOn.Format(time.ANSIC) + ")_"
		}
		for _, p := range q.SpokenBy {
			str += "\n>\t- " + p.Name
		}
		return str
	}

	display := func(quotes []contract.Quote, err error, channel string) {
		if err != nil {
			reportErr(err, channel)
		}

		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Found %v Quotes", len(quotes)), channel))

		for i := 0; i < int(math.Min(10, float64(len(quotes)))); i++ {
			rtm.SendMessage(rtm.NewOutgoingMessage(formatQuote(quotes[i]), channel))
		}
	}

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				info := rtm.GetInfo()

				text := ev.Text
				text = strings.TrimSpace(text)

				if ev.User != info.User.ID {
					config.Logger.Debug(text)

					listAllMatch, _ := regexp.MatchString("<@UGU8HMXC5> list all", text)
					if listAllMatch {
						quotes, err := dao.QuoteDao.List(nil)
						display(quotes, err, ev.Channel)
					}

					searchMatch, _ := regexp.MatchString("<@UGU8HMXC5> search", text)
					if searchMatch {
						query := strings.TrimPrefix(text, "<@UGU8HMXC5> search ")
						config.Logger.Debugw("searching",
							"QUERY", query,
						)

						quotes, err := dao.QuoteDao.Search(query, nil)

						display(quotes, err, ev.Channel)
					}

					createMatch, _ := regexp.MatchString("<@UGU8HMXC5> create", text)
					if createMatch {
						text = strings.TrimPrefix(text, "create")
						r, err := regexp.Compile(`"[^"]+"`)

						if err != nil {
							reportErr(err, ev.Channel)
						}

						matches := r.FindAllString(text, 10)
						if matches == nil || len(matches) < 2 {
							displayErr(errors.New(`Could not parse quote text, format: "[QUOTE]"( "[PERSON]")+ `), ev.Channel)
						}

						now := time.Now()
						quote := contract.Quote{
							Text:      strings.Trim(matches[0], "\""),
							SpokenBy:  []contract.Person{},
							CreatedOn: &now,
						}

						for i := 1; i < len(matches); i++ {
							quote.SpokenBy = append(quote.SpokenBy, contract.Person{Name: strings.Trim(matches[i], "\"")})
						}

						err = dao.QuoteDao.Create(quote)
						if err != nil {
							reportErr(err, ev.Channel)
						}

						rtm.SendMessage(rtm.NewOutgoingMessage("New Quote Created!\n"+formatQuote(quote), ev.Channel))
					}
				}
			case *slack.RTMError:
				config.Logger.Errorf("Error: %s\n", ev.Error())
			case *slack.InvalidAuthEvent:
				config.Logger.Error("Invalid credentials")
			default:
				// Take no action
			}
		}
	}
}
