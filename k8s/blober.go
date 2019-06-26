package k8s

import (
	"code.cloudfoundry.org/lager"
	"context"
	"github.com/concourse/go-archive/tgzfs"
	"os"
	"fmt"
)

const (
	BUCKET_TYPE_S3 = "s3"
	BUCKET_TYPE_GCS = "gcs"
)

func main(){
	bucketConfig := BucketConfig{
		Type: "s3",
		URL: "s3://concourse-rootfs",
		Secret: "notasecret",
	}
	Pull(lager.NewLogger("testpull"), context.TODO(), bucketConfig, "rootfs.tar.gz", "/tmp")
}

type BucketConfig struct {
	Type string
	URL string
	// if s3 should inclue AWS_ACCESS_KEY and AWS_SECRET and AWS_REGION,
	// if GCS should include GCP_JSON_FILE_PATH
	Secret string
}

func Pull(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourceKey, destinationPath string) error {
	blob := NewBlobReaderWriter(getBucketURL(bucket), sourceKey, destinationPath)

	blobReader, err:= blob.InputBlobReader(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-reader", err)
		return err
	}

	tgzfs.Extract(blobReader, destinationPath)
	return nil
}

func Push(logger lager.Logger, ctx context.Context, bucket BucketConfig, sourcePath, destinationKey string) error {
	blob := NewBlobReaderWriter(getBucketURL(bucket), sourcePath, destinationKey)

	blobWriter, err:= blob.OutputBlobWriter(logger, ctx)
	if err != nil {
		logger.Error("could-not-create-blob-writer", err)
		return err
	}

	tgzfs.Compress(blobWriter, destinationKey)
	return nil
}

func getBucketURL(bucket BucketConfig) string {
	url := bucket.URL
	if bucket.Type == BUCKET_TYPE_S3 {
		region := os.Getenv("AWS_REGION")
		url = fmt.Sprintf("%s/?region=%s", url, region)
	}
	return url
}
