package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	Client  *minio.Client
	buckets []string
}

func New(ctx context.Context, endpoint, accessKey, secretKey string, useSSL bool, buckets []string) (*MinIO, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	m := &MinIO{
		Client:  client,
		buckets: buckets,
	}

	if err := m.ensureBuckets(ctx); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MinIO) ensureBuckets(ctx context.Context) error {
	for _, bucket := range m.buckets {
		exists, err := m.Client.BucketExists(ctx, bucket)
		if err != nil {
			return err
		}

		if !exists {
			if err := m.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}
