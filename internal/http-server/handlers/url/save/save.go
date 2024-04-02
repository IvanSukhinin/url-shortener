package save

import (
	"context"
	"errors"
	"io"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/middleware/auth"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
)

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" validate:"omitempty,alphanum"`
}

type AliasSaver interface {
	SaveAlias(url string, alias string) error
}

func New(cfg *config.Config, log *slog.Logger, aliasSaver AliasSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

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

		var req Request
		err = render.DecodeJSON(r.Body, &req)
		// io.EOF - get an empty body
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			render.JSON(w, r, resp.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// validate request data
		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			var validatorErr validator.ValidationErrors
			errors.As(err, &validatorErr)

			render.JSON(w, r, resp.ValidationError(validatorErr))
			return
		}

		// fill alias
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(cfg.AliasLength)
		}

		// try to save url alias
		err = aliasSaver.SaveAlias(req.Url, alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasExists) {
				log.Info("alias already exists", slog.String("alias", req.Alias))
				render.JSON(w, r, resp.Error("alias already exists"))
				return
			}
			log.Error("failed to add alias", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add alias"))
			return
		}

		render.JSON(w, r, resp.OK())
	}
}
