package game

import (
	"math/rand"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
)

const CallDeckShuffleVersion = "v1"

func BuildCallDeck(words []domain.WordSetWord, seed, version string) []domain.WordSetWord {
	shuffled := append([]domain.WordSetWord(nil), words...)
	random := rand.New(rand.NewSource(seedInt64(seed + ":" + version)))
	random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}
