package handlers

import (
	"context"
	"net/http"

	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/bytedance/gopkg/util/logger"
)

type Generator interface {
	Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error)
}

type Server struct {
	gen Generator
}

func NewServer(gen Generator) *Server {
	return &Server{gen: gen}
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

	wod, err := server.gen.Generate(ctx, string(req.Body.Level), req.Body.DurationMin, equipment, req.Body.Seed)
	if err != nil {
		logger.Error("server.gen.Generate failed", "err", err)

		return &GenerateWod400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		}, nil
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
