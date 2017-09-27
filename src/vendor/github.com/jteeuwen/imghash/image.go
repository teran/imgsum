// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imghash

import (
	"image"
	"image/color"
)

// grayscale turns the image into a grayscale image.
func grayscale(img image.Image) image.Image {
	rect := img.Bounds()
	gray := image.NewGray(rect)

	var x, y int
	for y = rect.Min.Y; y < rect.Max.Y; y++ {
		for x = rect.Min.X; x < rect.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	return gray
}

// average converts the sums to averages and returns the result.
func average(sum []uint64, w, h int, n uint64) image.Image {
	ret := image.NewRGBA(image.Rect(0, 0, w, h))
	pix := ret.Pix

	var x, y, idx int
	for y = 0; y < h; y++ {
		for x = 0; x < w; x++ {
			idx = 4 * (y*w + x)
			pix[idx] = uint8(sum[idx] / n)
			pix[idx+1] = uint8(sum[idx+1] / n)
			pix[idx+2] = uint8(sum[idx+2] / n)
			pix[idx+3] = uint8(sum[idx+3] / n)
		}
	}

	return ret
}

// resize returns a scaled copy of the image slice r of m.
// The returned image has width w and height h.
func resize(m image.Image, w, h int) image.Image {
	if w < 0 || h < 0 {
		return nil
	}

	r := m.Bounds()
	if w == 0 || h == 0 || r.Dx() <= 0 || r.Dy() <= 0 {
		return image.NewRGBA64(image.Rect(0, 0, w, h))
	}

	switch m := m.(type) {
	case *image.RGBA:
		return resizeRGBA(m, r, w, h)

	case *image.YCbCr:
		if m, ok := resizeYCbCr(m, r, w, h); ok {
			return m
		}
	}

	ww, hh := uint64(w), uint64(h)
	dx, dy := uint64(r.Dx()), uint64(r.Dy())
	n, sum := dx*dy, make([]uint64, 4*w*h)

	var x, y int
	var r32, g32, b32, a32 uint32
	var r64, g64, b64, a64, remx, remy, index uint64
	var py, px, qx, qy uint64

	minx, miny := r.Min.X, r.Min.Y
	maxx, maxy := r.Max.X, r.Max.Y

	for y = miny; y < maxy; y++ {
		for x = minx; x < maxx; x++ {
			// Get the source pixel.
			r32, g32, b32, a32 = m.At(x, y).RGBA()
			r64 = uint64(r32)
			g64 = uint64(g32)
			b64 = uint64(b32)
			a64 = uint64(a32)

			// Spread the source pixel over 1 or more destination rows.
			py = uint64(y) * hh
			for remy = hh; remy > 0; {
				qy = dy - (py % dy)

				if qy > remy {
					qy = remy
				}

				// Spread the source pixel over 1 or more destination columns.
				px = uint64(x) * ww
				index = 4 * ((py/dy)*ww + (px / dx))

				for remx = ww; remx > 0; {
					qx = dx - (px % dx)

					if qx > remx {
						qx = remx
					}

					sum[index] += r64 * qx * qy
					sum[index+1] += g64 * qx * qy
					sum[index+2] += b64 * qx * qy
					sum[index+3] += a64 * qx * qy
					index += 4
					px += qx
					remx -= qx
				}

				py += qy
				remy -= qy
			}
		}
	}

	return average(sum, w, h, n*0x0101)
}

// resizeYCbCr returns a scaled copy of the YCbCr image slice r of m.
// The returned image has width w and height h.
func resizeYCbCr(m *image.YCbCr, r image.Rectangle, w, h int) (image.Image, bool) {
	var verticalRes int

	switch m.SubsampleRatio {
	case image.YCbCrSubsampleRatio420:
		verticalRes = 2
	case image.YCbCrSubsampleRatio422:
		verticalRes = 1
	default:
		return nil, false
	}

	ww, hh := uint64(w), uint64(h)
	dx, dy := uint64(r.Dx()), uint64(r.Dy())
	n, sum := dx*dy, make([]uint64, 4*w*h)

	var x, y int
	var r8, g8, b8 uint8
	var r64, g64, b64, remx, remy, index uint64
	var py, px, qx, qy, qxy uint64
	var Y, Cb, Cr []uint8

	minx, miny := r.Min.X, r.Min.Y
	maxx, maxy := r.Max.X, r.Max.Y

	for y = miny; y < maxy; y++ {
		Y = m.Y[y*m.YStride:]
		Cb = m.Cb[y/verticalRes*m.CStride:]
		Cr = m.Cr[y/verticalRes*m.CStride:]

		for x = minx; x < maxx; x++ {
			// Get the source pixel.
			r8, g8, b8 = color.YCbCrToRGB(Y[x], Cb[x/2], Cr[x/2])
			r64 = uint64(r8)
			g64 = uint64(g8)
			b64 = uint64(b8)

			// Spread the source pixel over 1 or more destination rows.
			py = uint64(y) * hh

			for remy = hh; remy > 0; {
				qy = dy - (py % dy)

				if qy > remy {
					qy = remy
				}

				// Spread the source pixel over 1 or more destination columns.
				px = uint64(x) * ww
				index = 4 * ((py/dy)*ww + (px / dx))

				for remx = ww; remx > 0; {
					qx = dx - (px % dx)

					if qx > remx {
						qx = remx
					}

					qxy = qx * qy
					sum[index] += r64 * qxy
					sum[index+1] += g64 * qxy
					sum[index+2] += b64 * qxy
					sum[index+3] += 0xFFFF * qxy
					index += 4
					px += qx
					remx -= qx
				}

				py += qy
				remy -= qy
			}
		}
	}

	return average(sum, w, h, n), true
}

// resizeRGBA returns a scaled copy of the RGBA image slice r of m.
// The returned image has width w and height h.
func resizeRGBA(m *image.RGBA, r image.Rectangle, w, h int) image.Image {
	ww, hh := uint64(w), uint64(h)
	dx, dy := uint64(r.Dx()), uint64(r.Dy())
	n, sum := dx*dy, make([]uint64, 4*w*h)

	var x, y int
	var pixOffset int
	var r64, g64, b64, a64, remx, remy, index uint64
	var py, px, qx, qy, qxy uint64

	minx, miny := r.Min.X, r.Min.Y
	maxx, maxy := r.Max.X, r.Max.Y

	for y = miny; y < maxy; y++ {
		pixOffset = m.PixOffset(minx, y)

		for x = minx; x < maxx; x++ {
			// Get the source pixel.
			r64 = uint64(m.Pix[pixOffset+0])
			g64 = uint64(m.Pix[pixOffset+1])
			b64 = uint64(m.Pix[pixOffset+2])
			a64 = uint64(m.Pix[pixOffset+3])
			pixOffset += 4

			// Spread the source pixel over 1 or more destination rows.
			py = uint64(y) * hh

			for remy = hh; remy > 0; {
				qy = dy - (py % dy)

				if qy > remy {
					qy = remy
				}

				// Spread the source pixel over 1 or more destination columns.
				px = uint64(x) * ww
				index = 4 * ((py/dy)*ww + (px / dx))

				for remx = ww; remx > 0; {
					qx = dx - (px % dx)

					if qx > remx {
						qx = remx
					}

					qxy = qx * qy
					sum[index] += r64 * qxy
					sum[index+1] += g64 * qxy
					sum[index+2] += b64 * qxy
					sum[index+3] += a64 * qxy
					index += 4
					px += qx
					remx -= qx
				}

				py += qy
				remy -= qy
			}
		}
	}

	return average(sum, w, h, n)
}
