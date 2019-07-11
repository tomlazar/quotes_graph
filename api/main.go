package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomlazar/quotes_graph/api/controllers"

	"github.com/tomlazar/quotes_graph/config"
	"github.com/tomlazar/quotes_graph/dao"
)

func main() {
	log := config.NewLogger("api")
	log.Debug("API starting")

	d, err := dao.NewDao()
	if err != nil {
		log.Panic(err)
	}

	log.Info("API started succesfully")

	router := controllers.Routes(d, log)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Debug(fmt.Sprintf("%s %s\n", method, route))
		return nil
	}

	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":8080", router))
}
