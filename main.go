package main

import (
	"context"

	_ "github.com/tomlazar/quotes_graph/config"

	"github.com/tomlazar/quotes_graph/commands"
	"github.com/tomlazar/quotes_graph/dao"
)

func main() {
	dao, err := dao.NewDao()
	if err != nil {
		panic(err)
	}

	bot := commands.NewBot(dao)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = bot.Listen(ctx)
	if err != nil {
		panic(err)
	}
}
