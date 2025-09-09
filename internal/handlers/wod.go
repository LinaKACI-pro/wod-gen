package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/LinaKACI-pro/wod-gen/internal/core"
	"github.com/bytedance/gopkg/util/logger"
)

type Server struct {
	wodGenerate core.WodGeneratorInterface
	wodList     core.WodListInterface
}

func NewServer(wodGenerate core.WodGeneratorInterface, list core.WodListInterface) *Server {
	return &Server{wodGenerate: wodGenerate, wodList: list}
}

func (server *Server) GenerateWod(ctx context.Context, req GenerateWodRequestObject) (GenerateWodResponseObject, error) {
	if req.Body == nil {
		return &GenerateWod400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: "missing body",
		}, nil
	}

	var equipment []string
	if req.Body.Equipment != nil {
		equipment = *req.Body.Equipment
	} else {
		equipment = nil
	}

	wod, err := server.wodGenerate.Generate(ctx, string(req.Body.Level), req.Body.DurationMin, equipment, req.Body.Seed)
	if err != nil {
		logger.Error("server.wodGenerate.Generate()", slog.Any("err", err))

		var invalidDataErr common.InvalidDataError
		switch {
		case errors.As(err, &invalidDataErr):
			return &GenerateWod400JSONResponse{
				Code:    http.StatusBadRequest,
				Message: invalidDataErr.Error(),
			}, nil
		case errors.Is(err, common.ErrDuration),
			errors.Is(err, common.ErrEmptyCatalog),
			errors.Is(err, common.ErrNoMoves):
			return &GenerateWod400JSONResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		default:
			return &GenerateWod500JSONResponse{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			}, nil
		}
	}

	blocks := make([]Block, len(wod.Blocks))
	for i, b := range wod.Blocks {
		blocks[i] = Block{Name: &b.Name, Params: &b.Params}
	}

	return &GenerateWod200JSONResponse{
		Blocks:           blocks,
		CreatedAt:        wod.CreatedAt,
		DurationMin:      wod.DurationMin,
		Equipment:        &wod.Equipment,
		GeneratorVersion: "v1",
		Id:               wod.ID,
		Level:            WodLevel(wod.Level),
		Seed:             wod.Seed,
	}, nil
}

func (server *Server) ListWods(ctx context.Context, req ListWodsRequestObject) (ListWodsResponseObject, error) {
	limit := 10
	offset := 0
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}

	wods, err := server.wodList.List(ctx, limit, offset)
	if err != nil {
		logger.Error("server.wodList.List()", slog.Any("err", err))
		return &ListWods500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to list wods",
		}, nil
	}

	resp := make([]Wod, len(wods))
	for i, w := range wods {
		blocks := make([]Block, len(w.Blocks))
		for j, b := range w.Blocks {
			blocks[j] = Block{Name: &b.Name, Params: &b.Params}
		}
		resp[i] = Wod{
			Id:               w.ID,
			Seed:             w.Seed,
			CreatedAt:        w.CreatedAt,
			Level:            WodLevel(w.Level),
			DurationMin:      w.DurationMin,
			Equipment:        &w.Equipment,
			Blocks:           blocks,
			GeneratorVersion: "v1",
		}
	}

	return &ListWods200JSONResponse{Wods: &resp}, nil
}
