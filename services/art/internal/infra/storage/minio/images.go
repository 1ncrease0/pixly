package minio

import "github.com/minio/minio-go/v7"

type ImageProvider struct {
	client *minio.Client
}

func NewImageProvider(client *minio.Client) ImageProvider {
	return ImageProvider{
		client: client,
	}
}
