package minio

import (
	"context"
	"github.com/1ncrease0/pixly/services/art/internal/domain"
	"github.com/minio/minio-go/v7"
	"io"
	"net/url"
	"time"
)

type ImageProvider struct {
	client     *minio.Client
	bucket     string
	presignTTL time.Duration
}

func NewImageProvider(client *minio.Client, bucket string, presignTTL time.Duration) ImageProvider {
	return ImageProvider{
		client:     client,
		bucket:     bucket,
		presignTTL: presignTTL,
	}
}

func (p *ImageProvider) GetImageURL(ctx context.Context, objectKey string) (string, error) {
	_, err := p.client.StatObject(ctx, p.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return "", domain.ErrImageNotFound
	}
	reqParams := make(url.Values)
	u, err := p.client.PresignedGetObject(
		ctx,
		p.bucket,
		objectKey,
		p.presignTTL,
		reqParams,
	)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (p *ImageProvider) SaveImage(ctx context.Context, objectKey string, reader io.Reader, size int64) error {
	_, err := p.client.PutObject(
		ctx,
		p.bucket,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "image/png",
		},
	)
	return err
}

func (p *ImageProvider) DeleteImage(ctx context.Context, objectKey string) error {
	_, err := p.client.StatObject(ctx, p.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return domain.ErrImageNotFound
	}

	return p.client.RemoveObject(
		ctx,
		p.bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	)
}
