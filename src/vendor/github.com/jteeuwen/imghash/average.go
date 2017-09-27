// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

package imghash

import "image"

// Average computes a Perceptual Hash using a naive, but very fast method.
// It holds up to minor colour changes, changing brightness and contrast and
// is indifferent to aspect ratio and image size differences.
//
// Average Hash is a great algorithm if you are looking for something specific.
// For example, if we have a small thumbnail of an image and we wish to know
// if the big one exists somewhere in our collection. Average Hash will find
// it very quickly. However, if there are modifications -- like text was added
// or a head was spliced into place, then Average Hash probably won't do the job.
//
// The Average Hash is quick and easy, but it can generate false-misses if
// gamma correction or color histogram is applied to the image. This is
// because the colors move along a non-linear scale -- changing where the
// "average" is located and therefore changing which bits are above/below the
// average.
func Average(img image.Image) uint64 {
	img = resize(img, 8, 8)
	img = grayscale(img)
	mean := avgMean(img)
	return avgHash(img, mean)
}

// avgMean computes the mean of all pixels.
func avgMean(img image.Image) uint32 {
	var x, y int
	var r, m uint32

	rect := img.Bounds()
	w := rect.Max.X - rect.Min.X
	h := rect.Max.Y - rect.Min.Y
	c := uint32(w * h)

	if c == 0 {
		return 0
	}

	for y = rect.Min.Y; y < rect.Max.Y; y++ {
		for x = rect.Min.X; x < rect.Max.X; x++ {
			r, _, _, _ = img.At(x, y).RGBA()
			m += r
		}
	}

	return m / c
}

// avgHash computes the hash bits for the given image and mean.
// It sets individual bits in a 64-bit integer. A bit is set if the
// pixel value is larger than the mean.
func avgHash(img image.Image, mean uint32) uint64 {
	var x, y int
	var value, bit uint64
	var r uint32

	rect := img.Bounds()

	for y = rect.Min.Y; y < rect.Max.Y; y++ {
		for x = rect.Min.X; x < rect.Max.X; x++ {
			r, _, _, _ = img.At(x, y).RGBA()

			if r > mean {
				value |= 1 << bit
			}

			bit++
		}
	}

	return value
}
