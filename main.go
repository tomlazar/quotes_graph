package main

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/getsentry/raven-go"

	_ "github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/contract"
	"github.com/tomlazar/quotes_graph/dao"

	"github.com/Sirupsen/logrus"
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
		logrus.Fatalln(err)
	}

	reportErr := func(err error, channel string) {
		raven.CaptureError(err, nil)
		logrus.Errorln(err)
		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Could not execute: %v", err), channel))
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
				text = strings.TrimPrefix(text, "<@UGU8HMXC5> ")

				if ev.User != info.User.ID {
					logrus.Debugln(text)

					listAllMatch, _ := regexp.MatchString("list all", text)
					if listAllMatch {
						quotes, err := dao.QuoteDao.List(nil)
						display(quotes, err, ev.Channel)
					}

					searchMatch, _ := regexp.MatchString("search", text)
					if searchMatch {
						query := strings.TrimPrefix(text, "search ")
						logrus.WithField("QUERY", query).Debugln("searching")
						quotes, err := dao.QuoteDao.Search(query, nil)

						display(quotes, err, ev.Channel)
					}

					createMatch, _ := regexp.MatchString("create", text)
					if createMatch {
						text = strings.TrimPrefix(text, "create")
						r, err := regexp.Compile(`"([^"]*)"`)

						logrus.Debugln("in create")

						if err != nil {
							reportErr(err, ev.Channel)
						}

						matches := r.FindAllStringSubmatch(text, 10)
						if matches == nil {
							reportErr(errors.New(`Could not parse quote text, format: "[QUOTE]"( "[PERSON]")+ `), ev.Channel)
						}

						now := time.Now()
						quote := contract.Quote{
							Text:      matches[0][1],
							SpokenBy:  []contract.Person{},
							CreatedOn: &now,
						}

						for i := 1; i < len(matches); i++ {
							quote.SpokenBy = append(quote.SpokenBy, contract.Person{Name: matches[i][1]})
						}

						err = dao.QuoteDao.Create(quote)
						if err != nil {
							reportErr(err, ev.Channel)
						}

						rtm.SendMessage(rtm.NewOutgoingMessage("New Quote Created!"+formatQuote(quote), ev.Channel))
					}
				}

			case *slack.RTMError:
				logrus.Errorf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				logrus.Fatalln("Invalid credentials")

			default:
				// Take no action
			}
		}
	}
}
