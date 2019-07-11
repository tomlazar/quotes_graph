package controllers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/tomlazar/quotes_graph/dao"
	"go.uber.org/zap"
)

// Routes creates all the controllers and the routes
func Routes(d *dao.Dao, log *zap.SugaredLogger) *chi.Mux {
	r := chi.NewRouter()

	// middlewear
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		zapMiddle(log),
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	// routes
	r.Route("/v1", func(r chi.Router) {
		r.Mount("/quotes", QuotesController(d, log))
	})
	return r
}

func zapMiddle(l *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				l.Infow("Served",
					"proto", r.Proto,
					"proto", r.Proto,
					"path", r.URL.Path,
					"lat", time.Since(t1),
					"status", ww.Status(),
					"size", ww.BytesWritten(),
					"reqId", middleware.GetReqID(r.Context()))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
