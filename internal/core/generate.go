package core

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/common"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	Beginner     string = "beginner"
	Intermediate string = "intermediate"
	Advanced     string = "advanced"
	MinDuration         = 15
	MaxDuration         = 120
)

type rng [2]int

type move struct {
	Name       string                    `yaml:"name"`
	NeedsOneOf []string                  `yaml:"needs_one_of"`
	Tags       []string                  `yaml:"tags"`
	Weight     float64                   `yaml:"weight"`
	Ranges     map[string]map[string]rng `yaml:"ranges"` // level -> param -> [min,max]
}

type Catalog struct {
	Moves []move `yaml:"moves"`
}

type Params struct {
	Level       string
	DurationMin int
	Equipment   []string
	Seed        string
}

func NewCatalog(raw []byte) (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	for i := range c.Moves {
		if c.Moves[i].Weight == 0 {
			c.Moves[i].Weight = 1.0
		}
	}

	return &Catalog{Moves: c.Moves}, nil
}

func (c *Catalog) Generate(ctx context.Context, level string, durationMin int, equipment []string, seed *string) (models.Wod, error) {
	_ = ctx
	lv := strings.ToLower(level)
	if lv != Beginner && lv != Intermediate && lv != Advanced {
		return models.Wod{}, common.InvalidDataError{DataType: "level", Data: lv}
	}

	if durationMin < MinDuration || durationMin > MaxDuration {
		return models.Wod{}, common.ErrDuration
	}

	if len(c.Moves) == 0 {
		return models.Wod{}, common.ErrEmptyCatalog
	}

	var parsedSeed string
	if seed != nil && *seed != "" {
		parsedSeed = *seed
	} else {
		parsedSeed = uuid.NewString()
	}
	hs := seedHash(parsedSeed, durationMin, lv, equipment)
	rnd := rand.New(rand.NewSource(hs)) //nolint:gosec // G404: deterministic, non-crypto PRNG is intended for workout sampling

	avail := filterByEquipment(c.Moves, equipment)
	if len(avail) == 0 {
		avail = filterNoEquipment(c.Moves)
		if len(avail) == 0 {
			return models.Wod{}, common.ErrNoMoves
		}
	}

	n := blocksForDuration(lv, durationMin)

	blocks := make([]models.Block, 0, n)
	var last string
	for i := 0; i < n; i++ {
		m := weightedPick(rnd, avail, last)
		last = m.Name
		params := pickParams(rnd, m.Ranges[lv])
		blocks = append(blocks, models.Block{Name: m.Name, Params: params})
	}

	return models.Wod{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		Level:       lv,
		DurationMin: durationMin,
		Equipment:   cloneStrings(equipment),
		Seed:        parsedSeed,
		Blocks:      blocks,
	}, nil
}
