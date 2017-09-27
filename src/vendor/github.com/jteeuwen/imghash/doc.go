// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

/*
imghash computes the Perceptual Hash for a given input image.
The Perceptual Hash is returned as a 64 bit integer.

Comparing two images can be done by constructing the hash from each image
and counting the number of bit positions that are different. This is a
Hamming distance. A distance of zero indicates that it is likely a very
similar picture (or a variation of the same picture). A distance of 5 means a
few things may be different, but they are probably still close enough to be
similar. But a distance of 10 or more? That's probably a very different picture.
*/
package imghash
