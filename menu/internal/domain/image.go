package domain

import "io"

type ImageContentType string

const (
	ImageContentTypeJPEG = "image/jpeg"
	ImageContentTypePNG  = "image/png"
	ImageContentTypeWebP = "image/webp"
)

type Image struct {
	Data        io.ReadCloser
	ContentType ImageContentType
}

func (i *Image) Close() error {
	return i.Data.Close()
}
