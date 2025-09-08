package core

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/LinaKACI-pro/wod-gen/internal/core/catalog"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/LinaKACI-pro/wod-gen/internal/repository"
	"github.com/google/uuid"
)

const (
	Beginner     string = "beginner"
	Intermediate string = "intermediate"
	Advanced     string = "advanced"
	MinDuration         = 15
	MaxDuration         = 120
)

type Params struct {
	Level       string
	DurationMin int
	Equipment   []string
	Seed        string
}

type WodGeneratorInterface interface {
	Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error)
}

type WodGenerator struct {
	wodRepository repository.WodRepositoryInterface
	catalog       *catalog.Catalog
}

func NewWodGenerator(catalog *catalog.Catalog, wodRepository repository.WodRepositoryInterface) *WodGenerator {
	return &WodGenerator{catalog: catalog, wodRepository: wodRepository}
}

func (w *WodGenerator) Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error) {
	lv, parsedSeed, err := validateInfo(level, durationMin, seed, w.catalog.Moves)
	if err != nil {
		return models.Wod{}, fmt.Errorf("validateInfo(): %w", err)
	}

	wod, err := buildWod(lv, durationMin, equipment, parsedSeed, w.catalog.Moves)
	if err != nil {
		return models.Wod{}, fmt.Errorf("buildWod(): %w", err)
	}

	savedWod, err := w.wodRepository.SaveWod(ctx, wod)
	if err != nil {
		return models.Wod{}, fmt.Errorf("wodRepository.SaveWod(): %w", err)
	}

	return savedWod, nil
}

func validateInfo(level string, durationMin int, seed *string, moves []catalog.Move) (string, string, error) {
	lv := strings.ToLower(level)
	if lv != Beginner && lv != Intermediate && lv != Advanced {
		return "", "", common.InvalidDataError{DataType: "level", Data: lv}
	}

	if durationMin < MinDuration || durationMin > MaxDuration {
		return "", "", common.ErrDuration
	}

	if len(moves) == 0 {
		return "", "", common.ErrEmptyCatalog
	}

	var parsedSeed string
	if seed != nil && *seed != "" {
		parsedSeed = *seed
	} else {
		parsedSeed = uuid.NewString()
	}

	return lv, parsedSeed, nil
}

func buildWod(level string, durationMin int, equipment []string, parsedSeed string, moves []catalog.Move) (models.Wod, error) {
	hs := seedHash(parsedSeed, durationMin, level, equipment)
	rnd := rand.New(rand.NewSource(hs)) //nolint:gosec // deterministic

	avail := filterByEquipment(moves, equipment)
	if len(avail) == 0 {
		avail = filterNoEquipment(moves)
		if len(avail) == 0 {
			return models.Wod{}, common.ErrNoMoves
		}
	}

	n := blocksForDuration(level, durationMin)

	blocks := make([]models.Block, 0, n)
	var last string
	for i := 0; i < n; i++ {
		m := weightedPick(rnd, avail, last)
		last = m.Name
		params := pickParams(rnd, m.Ranges[level])
		blocks = append(blocks, models.Block{Name: m.Name, Params: params})
	}

	return models.Wod{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		Level:       level,
		DurationMin: durationMin,
		Equipment:   cloneStrings(equipment),
		Seed:        parsedSeed,
		Blocks:      blocks,
	}, nil
}
