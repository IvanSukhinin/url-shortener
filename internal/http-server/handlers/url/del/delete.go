package del

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/http-server/middleware/auth"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
)

type AliasDeleter interface {
	DeleteAlias(alias string) error
}

func New(log *slog.Logger, aliasDeleter AliasDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// check rules
		err, _ := auth.ErrorFromContext(r.Context())
		if err != nil {
			log.Error(err.Error())

			ctx := context.WithValue(r.Context(), render.StatusCtxKey, http.StatusForbidden)
			render.JSON(w, r.WithContext(ctx), resp.Error(err.Error()))

			return
		}
		isAdmin, _ := auth.IsAdminFromContext(r.Context())
		if !isAdmin {
			log.Error("user is not admin. operation does not permitted")

			ctx := context.WithValue(r.Context(), render.StatusCtxKey, http.StatusForbidden)
			render.JSON(w, r.WithContext(ctx), resp.Error("operation does not permitted"))

			return
		}

		params := r.URL.Query()
		alias := params.Get("alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		err = aliasDeleter.DeleteAlias(alias)
		if err != nil {
			log.Error("failed to delete alias", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("alias deleted", slog.String("alias", alias))

		render.JSON(w, r, http.StatusOK)
	}
}
