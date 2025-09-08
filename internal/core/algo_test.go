package core

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlocksForDuration(t *testing.T) {
	require.Equal(t, 4, blocksForDuration(Beginner, 20))
	require.Equal(t, 6, blocksForDuration(Beginner, 40))
	require.Equal(t, 7, blocksForDuration(Beginner, 60))

	require.Equal(t, 6, blocksForDuration(Advanced, 25))
	require.Equal(t, 8, blocksForDuration(Advanced, 45))
	require.Equal(t, 10, blocksForDuration(Advanced, 90))

	require.Equal(t, 5, blocksForDuration(Intermediate, 20))
	require.Equal(t, 7, blocksForDuration(Intermediate, 40))
	require.Equal(t, 8, blocksForDuration(Intermediate, 60))
}

func TestFilterByEquipment(t *testing.T) {
	list := []move{
		{Name: "Run"},
		{Name: "Row", NeedsOneOf: []string{"rower"}},
		{Name: "Bike", NeedsOneOf: []string{"bike"}},
	}

	// Avec rower → doit inclure Run + Row
	got := filterByEquipment(list, []string{"rower"})
	require.Len(t, got, 2)

	// Sans équipement → doit inclure seulement Run
	got = filterByEquipment(list, nil)
	require.Len(t, got, 1)
	require.Equal(t, "Run", got[0].Name)
}

func TestFilterNoEquipment(t *testing.T) {
	list := []move{
		{Name: "Run"},
		{Name: "Row", NeedsOneOf: []string{"rower"}},
	}
	got := filterNoEquipment(list)
	require.Len(t, got, 1)
	require.Equal(t, "Run", got[0].Name)
}

func TestWeightedPick(t *testing.T) {
	avail := []move{
		{Name: "A", Weight: 1},
		{Name: "B", Weight: 100}, // devrait sortir quasi toujours
	}
	rnd := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic non-crypto PRNG is intended
	pick := weightedPick(rnd, avail, "")
	require.Equal(t, "B", pick.Name)
}

func TestPickParams_Defaults(t *testing.T) {
	rnd := rand.New(rand.NewSource(1)) //nolint:gosec // deterministic non-crypto PRNG is intended

	// Cas vide → fallback reps=10
	got := pickParams(rnd, nil)
	require.Equal(t, defaultReps, got["reps"])

	// Cas avec bornes valides
	got = pickParams(rnd, map[string]rng{"reps": {5, 5}})
	require.Equal(t, 5, got["reps"])
}

func TestSeedHash_Deterministic(t *testing.T) {
	h1 := seedHash("seed", 30, Beginner, []string{"rower"})
	h2 := seedHash("seed", 30, Beginner, []string{"rower"})
	h3 := seedHash("other", 30, Beginner, []string{"rower"})

	require.Equal(t, h1, h2, "same inputs → same hash")
	require.NotEqual(t, h1, h3, "different inputs → different hash")
}

func TestCloneStrings(t *testing.T) {
	src := []string{"a", "b"}
	dst := cloneStrings(src)
	require.Equal(t, src, dst)
	require.NotSame(t, &src[0], &dst[0], "should copy underlying array")
}
