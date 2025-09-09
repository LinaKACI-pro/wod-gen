//nolint:gosec // deterministic PRNG is intended
package core

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/LinaKACI-pro/wod-gen/internal/core/catalog"
)

const (
	defaultReps     = 10
	repeatPenalty   = 0.1 // reduce weight if same catalog.Move as last
	minParamDefault = 1
)

func blocksForDuration(level string, d int) int {
	switch level {
	case Beginner:
		if d <= 25 {
			return 4
		} else if d <= 45 {
			return 6
		}
		return 7
	case Advanced:
		if d <= 30 {
			return 6
		} else if d <= 60 {
			return 8
		}
		return 10
	default: // Intermediate
		if d <= 30 {
			return 5
		} else if d <= 50 {
			return 7
		}
		return 8
	}
}

func filterByEquipment(list []catalog.Move, eqs []string) []catalog.Move {
	set := map[string]struct{}{}
	for _, s := range eqs {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			set[s] = struct{}{}
		}
	}
	out := make([]catalog.Move, 0, len(list))
	for _, m := range list {
		if len(m.NeedsOneOf) == 0 {
			out = append(out, m)
			continue
		}
		for _, need := range m.NeedsOneOf {
			if _, ok := set[strings.ToLower(need)]; ok {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

func filterNoEquipment(list []catalog.Move) []catalog.Move {
	out := make([]catalog.Move, 0, len(list))
	for _, m := range list {
		if len(m.NeedsOneOf) == 0 {
			out = append(out, m)
		}
	}
	return out
}

func weightedPick(rnd *rand.Rand, avail []catalog.Move, last string) catalog.Move {
	total := 0.0
	acc := make([]float64, len(avail))
	for i, m := range avail {
		w := m.Weight
		if m.Name == last {
			w *= repeatPenalty
		}
		total += w
		acc[i] = total
	}
	x := rnd.Float64() * total
	for i, a := range acc {
		if x <= a {
			return avail[i]
		}
	}
	// fallback: should not happen
	return avail[len(avail)-1]
}

func pickParams(rnd *rand.Rand, ranges map[string]catalog.Rng) map[string]interface{} {
	out := make(map[string]interface{}, len(ranges))
	if len(ranges) == 0 {
		out["reps"] = defaultReps
		return out
	}
	for k, mm := range ranges {
		minParam, maxParam := mm[0], mm[1]
		if minParam < minParamDefault {
			minParam = minParamDefault
		}
		if maxParam < minParam {
			maxParam = minParam
		}
		val := minParam
		if maxParam > minParam {
			val = minParam + rnd.Intn(maxParam-minParam+1)
		}
		if val < minParamDefault {
			val = minParamDefault
		}
		out[k] = val
	}
	if len(out) == 0 {
		out["reps"] = defaultReps
	}
	return out
}

func seedHash(seed string, dur int, level string, equip []string) int64 {
	h := sha256.New()
	_, err := fmt.Fprintf(h, "%s|%s|%d|%s",
		seed,
		level,
		dur,
		strings.Join(equip, ","),
	)
	if err != nil {
		return 0
	}
	sum := h.Sum(nil)
	u := binary.LittleEndian.Uint64(sum[:8]) & math.MaxInt64
	return int64(u)
}

func cloneStrings(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	return out
}
