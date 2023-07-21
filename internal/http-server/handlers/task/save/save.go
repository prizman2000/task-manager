package save

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"

	resp "lebedev.vr/task-manager/internal/lib/api/response"
	"lebedev.vr/task-manager/internal/lib/logger/sl"
	"lebedev.vr/task-manager/internal/storage"
)

type Request struct {
	Name string `json:"name" validate:"required"`
}

type Response struct {
	resp.Response
}

type TaskSaver interface {
	SaveTask(taskName string) (int64, error)
}

func New(log *slog.Logger, taskSaver TaskSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
            slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
            log.Error("failed to decode request body", sl.Err(err))
            
			render.JSON(w, r, resp.Err("failed to decode request body"))

            return
        }

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err!= nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request body", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
        }

		id, err := taskSaver.SaveTask(req.Name)
        if errors.Is(err, storage.ErrTaskExists) {
			log.Info("task with this name already exists", slog.String("name", req.Name))

			render.JSON(w, r, resp.Err("task with this name already exists"))

			return
		}
		if err != nil {
            log.Error("failed to save task", sl.Err(err))

			render.JSON(w, r, resp.Err("failed to save task"))

			return
		}

		log.Info("task saved successfully", slog.Int64("id", id))

		responseOk(w, r)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.Ok(),
	})
}