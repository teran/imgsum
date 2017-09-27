// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

package imghash

import "image"

// A HashFunc computes a Perceptual Hash for a given image.
type HashFunc func(image.Image) uint64

// Distance calculates the Hamming Distance between the two input hashes.
func Distance(a, b uint64) uint64 {
	var dist uint64

	for val := a ^ b; val != 0; val &= val - 1 {
		dist++
	}

	return dist
}
