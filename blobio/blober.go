package blobio

import (
	"code.cloudfoundry.org/lager"
	"context"
	"github.com/concourse/go-archive/tgzfs"
	"os"
	"path/filepath"
)

type BucketConfig struct {
	URL  string
	// if s3 should inclue AWS_ACCESS_KEY and AWS_SECRET and AWS_REGION,
	// if GCS should include GCP_JSON_FILE_PATH
}

func Pull(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourceKey, destinationPath string) error {
	blob := NewBlobReaderWriter(bucket.URL, sourceKey)

	blobReader, err := blob.InputBlobReader(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-reader", err)
		return err
	}

	return tgzfs.Extract(blobReader, destinationPath)

}

func Push(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourcePath, destinationKey string) error {
	blob := NewBlobReaderWriter(bucket.URL, destinationKey)

	blobWriter, err := blob.OutputBlobWriter(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-writer", err)
		return err
	}

	var paths []string
	err = filepath.Walk(sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() == false {
				paths = append(paths, path)
			}
			return nil
		})
	if err != nil {
		return err
	}

	return tgzfs.Compress(blobWriter, sourcePath, paths...)
}
