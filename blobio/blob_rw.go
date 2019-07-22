package blobio

import (
	"context"
	"io"

	"code.cloudfoundry.org/lager"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

//go:generate counterfeiter . BlobstoreIO
type BlobstoreIO interface {
	InputBlobReader(lager.Logger, context.Context) (io.Reader, error)
	OutputBlobWriter(lager.Logger, context.Context) (io.Writer, error)
}

type BlobReaderWriter struct {
	BucketURL  string
	BucketKey string

}

func NewBlobReaderWriter(bucketURL string, bucketKey string) BlobstoreIO {
	return &BlobReaderWriter{
		BucketURL:  bucketURL,
		BucketKey: bucketKey,
	}
}

func (brw *BlobReaderWriter) InputBlobReader(logger lager.Logger, ctx context.Context) (io.Reader, error) {
	bucket, err := blob.OpenBucket(ctx, brw.BucketURL)
	if err != nil {
		logger.Error("Failed to setup bucket: %s", err)
		return nil, err
	}
	defer bucket.Close()

	r, err := bucket.NewReader(ctx, brw.BucketKey, nil)
	if err != nil {
		logger.Error("Failed to obtain reader: %s", err)
		return nil, err
	}

	return r, nil
}

func (brw *BlobReaderWriter) OutputBlobWriter(logger lager.Logger, ctx context.Context) (io.Writer, error) {
	bucket, err := blob.OpenBucket(ctx, brw.BucketURL)
	if err != nil {
		logger.Error("Failed to setup bucket: %s", err)
		return nil, err
	}
	defer bucket.Close()

	w, err := bucket.NewWriter(ctx, brw.BucketKey, nil)
	if err != nil {
		logger.Error("Failed to obtain writer: %s", err)
		return nil, err
	}
	return w, nil
}
