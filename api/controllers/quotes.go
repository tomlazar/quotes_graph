package controllers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/tomlazar/quotes_graph/dao"
)

// QuotesController creates all the routes for the quotes api
func QuotesController(d *dao.Dao, log *zap.SugaredLogger) http.Handler {
	quotesLog := log.With("controller", "quotes")

	r := chi.NewRouter()
	r.Get("/", listQuotes(d, quotesLog))
	return r
}

func listQuotes(d *dao.Dao, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		skip := chi.URLParam(r, "skip")
		count := chi.URLParam(r, "count")

		rlog := log.With("path", r.URL, "skip", skip, "count", count)

		//TODO actaully use those vars
		list, err := d.QuoteDao.List(nil)
		if err != nil {
			rlog.Error(err)
		}

		render.JSON(w, r, list)
	}
}
