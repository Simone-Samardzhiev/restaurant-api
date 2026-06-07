package domain

import "io"

type ImageContentType string

const (
	ImageContentTypeJPEG ImageContentType = "image/jpeg"
	ImageContentTypePNG  ImageContentType = "image/png"
	ImageContentTypeWebP ImageContentType = "image/webp"
)

type Image struct {
	Data        io.ReadCloser
	ContentType ImageContentType
}

func (i *Image) Close() error {
	return i.Data.Close()
}
