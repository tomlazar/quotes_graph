package main

import (
	"github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/dao"
)

func main() {
	log := config.NewLogger("api")
	log.Debug("API starting")

	_, err := dao.NewDao()
	if err != nil {
		log.Panic(err)
	}

	log.Info("API started succesfully")
}
