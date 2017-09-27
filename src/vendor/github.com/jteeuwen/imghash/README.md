## imghash

imghash computes the Perceptual Hash for a given input image.
The hash is returned as a 64 bit integer. It comes with two commandline
tools: `img-index` and `img-find`. Refer to their respective READMEs for
information on what they do.

Note that this toolset is mainly for educational purposes on my part.
It is a partial implementation of an article on [hackerfactor.com][hf].

[hf]: http://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html

The package supports these hashing modes:

* **Average**: Average computes a Perceptual Hash using a naive, but very fast method.
  It holds up to minor colour changes, changing brightness and contrast and
  is indifferent to aspect ratio and image size differences.
  
  Average Hash is a great algorithm if you are looking for something specific.
  For example, if we have a small thumbnail of an image and we wish to know
  if the big one exists somewhere in our collection. Average Hash will find
  it very quickly. However, if there are modifications -- like text was added
  or a head was spliced into place, then Average Hash probably won't do the job.
  
  The Average Hash is quick and easy, but it can generate false-misses if
  gamma correction or color histogram is applied to the image. This is
  because the colors move along a non-linear scale -- changing where the
  "average" is located and therefore changing which bits are above/below the
  average.

More may come at some point.

### Usage

    go get github.com/jteeuwen/imghash


### License

Unless otherwise stated, all of the work in this project is subject to a
1-clause BSD license. Its contents can be found in the enclosed LICENSE file.

