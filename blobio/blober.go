package blobio

import (
	"context"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/go-archive/tarfs"
)

type BucketConfig struct {
	URL string
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

	err = tarfs.Extract(blobReader, destinationPath)
	if err != nil {
		return err
	}

	err = blobReader.Close()
	if err != nil {
		return err
	}
	return nil
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

	err = tarfs.Compress(blobWriter, sourcePath, paths...)
	if err != nil {
		return err
	}

	err = blobWriter.Close()
	if err != nil {
		return err
	}
	return nil
}
