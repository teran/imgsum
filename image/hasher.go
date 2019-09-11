package image

import (
	"fmt"

	"github.com/brett-lempereur/ish"
)

// HashType type to represent the list of hashers available
type HashType string

var (
	// HashTypeAvg reflects AverageHash
	HashTypeAvg HashType = "avg"

	// HashTypeDiff reflects DiffrenceHash
	HashTypeDiff HashType = "diff"
)

func getHasher(kind HashType, res int) (ish.PerceptualHash, error) {
	switch kind {
	case HashTypeAvg:
		return ish.NewAverageHash(res, res), nil
	case HashTypeDiff:
		return ish.NewDifferenceHash(res, res), nil
	default:
		return nil, fmt.Errorf("Unknown hasher: %s", kind)
	}
}
