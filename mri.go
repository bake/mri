package mri

import (
	"image"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/image/webp"
)

// DecodeConfig returns the color model and dimensions of a MRI. To recreate
// the original header it has to read the whole stream.
func DecodeConfig(r io.Reader) (image.Config, error) {
	r, _, err := RawDecode(r)
	if err != nil {
		return image.Config{}, err
	}
	return webp.DecodeConfig(r)
}

// Decode reads a MRI image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	r, _, err := RawDecode(r)
	if err != nil {
		return nil, err
	}
	return webp.Decode(r)
}

// RawDecode reads a MRI image from r, returns its original format (WEBP) as an
// io.Reader and its size in bytes. Can be used if there is no need for an
// image.Image.
func RawDecode(r io.Reader) (io.Reader, int, error) {
	bb, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, 0, errors.Wrap(err, "could not read image file")
	}
	pr, pw := io.Pipe()

	// Pretty much a rewrite of Tachiyomis' implementation.
	go func() {
		defer pw.Close()

		size := len(bb) + 7
		pw.Write([]byte("RIFF"))
		pw.Write([]byte{
			byte(size >> 0 & 255),
			byte(size >> 8 & 255),
			byte(size >> 16 & 255),
			byte(size >> 24 & 255),
		})
		pw.Write([]byte("WEBP"))
		pw.Write([]byte("VP8"))

		for _, b := range bb {
			pw.Write([]byte{b ^ 101})
		}
	}()

	headerSize := 15
	return pr, len(bb) + headerSize, nil
}

func init() {
	image.RegisterFormat("mri", "E", Decode, DecodeConfig)
}
