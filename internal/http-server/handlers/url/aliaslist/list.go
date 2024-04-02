package alist

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage/postgresql"
)

// AliasListGetter is an interface for getting list of aliases.
type AliasListGetter interface {
	GetAliasList() (*[]postgresql.Alias, error)
}

func New(log *slog.Logger, aliasListGetter AliasListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.alist.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		res, err := aliasListGetter.GetAliasList()
		if err != nil {
			log.Error(err.Error())
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("got aliases list")

		render.JSON(w, r, res)
	}
}
