package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getsentry/raven-go"

	"github.com/tomlazar/quotes_graph/contract"

	"github.com/tomlazar/quotes_graph/dao"

	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

func init() {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/quote_graph/")
	viper.AddConfigPath("$HOME/.quote_graph")
	viper.AddConfigPath(".")

	viper.SetDefault("interactive", true)
	viper.SetDefault("raven.dsn", nil)
	viper.SetDefault("debug", false)

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Warnln("Could not load in config file")
	}

	if viper.GetBool("interactive") {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	dsn := viper.Get("raven.dsn")
	if dsn != nil {
		raven.SetDSN(dsn.(string))
	}

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	token := viper.GetString("slack.token")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	dao, err := dao.NewDao()
	if err != nil {
		logrus.Fatalln(err)
	}

	display := func(quotes []contract.Quote, err error, channel string) {
		if err != nil {
			raven.CaptureError(err, nil)

			logrus.Errorln(err)
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Could not execute: %v", err), channel))
		}

		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Found %v Quotes", len(quotes)), channel))

		for _, q := range quotes {
			str := "> " + q.Text
			for _, p := range q.SpokenBy {
				str += "\n>\t- " + p
			}

			rtm.SendMessage(rtm.NewOutgoingMessage(str, channel))
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
				text = strings.ToLower(text)
				text = strings.TrimPrefix(text, "<@ugu8hmxc5> ")

				if ev.User != info.User.ID {
					logrus.Debugln(text)

					listAllMatch, _ := regexp.MatchString("list all", text)
					if listAllMatch {
						quotes, err := dao.QuoteDao.List(nil)
						display(quotes, err, ev.Channel)
					}

					searchMatch, _ := regexp.MatchString("^search", text)
					if searchMatch {
						query := strings.TrimPrefix(text, "search ")
						quotes, err := dao.QuoteDao.Search(query, nil)
						display(quotes, err, ev.Channel)
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
