package catalog

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Rng [2]int

type Move struct {
	Name       string                    `yaml:"name"`
	NeedsOneOf []string                  `yaml:"needs_one_of"`
	Tags       []string                  `yaml:"tags"`
	Weight     float64                   `yaml:"weight"`
	Ranges     map[string]map[string]Rng `yaml:"ranges"` // level -> param -> [min,max]
}

type Catalog struct {
	Moves []Move `yaml:"moves"`
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
