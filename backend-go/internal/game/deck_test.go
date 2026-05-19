package game

import (
	"reflect"
	"testing"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
)

func TestBuildCallDeckDeterministicFromSeedAndVersion(t *testing.T) {
	words := []domain.WordSetWord{
		{ID: "1", Word: "One"},
		{ID: "2", Word: "Two"},
		{ID: "3", Word: "Three"},
		{ID: "4", Word: "Four"},
		{ID: "5", Word: "Five"},
	}

	first := BuildCallDeck(words, "seed-a", CallDeckShuffleVersion)
	second := BuildCallDeck(words, "seed-a", CallDeckShuffleVersion)
	third := BuildCallDeck(words, "seed-b", CallDeckShuffleVersion)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected same seed/version to produce same deck, got %+v and %+v", first, second)
	}
	if reflect.DeepEqual(first, third) {
		t.Fatalf("expected different seed to produce different deck, got %+v", third)
	}
	if reflect.DeepEqual(words, first) {
		t.Fatalf("expected deck to be shuffled, got original order %+v", first)
	}
}
