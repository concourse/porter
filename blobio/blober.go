package blobio

import (
	"context"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/go-archive/tgzfs"
)

const (
	BUCKET_TYPE_S3  = "s3"
	BUCKET_TYPE_GCS = "gcs"
)


type BucketConfig struct {
	URL  string
	// if s3 should inclue AWS_ACCESS_KEY and AWS_SECRET and AWS_REGION,
	// if GCS should include GCP_JSON_FILE_PATH
	Secret string
	Region string
}

func (b BucketConfig) BucketType() string {
	//TODO
	return BUCKET_TYPE_S3
}

func Pull(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourceKey, destinationPath string) error {
	blob := NewBlobReaderWriter(bucket.URL, sourceKey, destinationPath)

	blobReader, err := blob.InputBlobReader(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-reader", err)
		return err
	}

	return tgzfs.Extract(blobReader, destinationPath)

}

func Push(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourcePath, destinationKey string) error {
	blob := NewBlobReaderWriter(getBucketURL(bucket), sourcePath, destinationKey)

	blobWriter, err := blob.OutputBlobWriter(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-writer", err)
		return err
	}

	return tgzfs.Compress(blobWriter, destinationKey)
}

func getBucketURL(bucket BucketConfig) string {
	url := bucket.URL
	if bucket.BucketType() == BUCKET_TYPE_S3 {
		region := os.Getenv("AWS_REGION")
		url = fmt.Sprintf("%s/?region=%s", url, region)
	}
	return url
}
