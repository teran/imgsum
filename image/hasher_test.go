package image

import (
	"fmt"
	"testing"

	"github.com/brett-lempereur/ish"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	r := require.New(t)

	type testCase struct {
		name      string
		ht        HashType
		expResult ish.PerceptualHash
		expError  error
	}

	tcs := []testCase{
		{
			name:      "average hash",
			ht:        HashTypeAvg,
			expResult: &ish.AverageHash{},
		},
		{
			name:      "difference hash",
			ht:        HashTypeDiff,
			expResult: &ish.DifferenceHash{},
		},
		{
			name:      "unexistent hash type",
			ht:        HashType("blah"),
			expResult: nil,
			expError:  fmt.Errorf("Unknown hasher: blah"),
		},
	}

	for _, tc := range tcs {
		hasher, err := getHasher(tc.ht, 1024)
		r.Equalf(tc.expError, err, tc.name)
		r.IsTypef(tc.expResult, hasher, tc.name)
	}
}
