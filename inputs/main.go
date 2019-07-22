package main

import (
	"context"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/porter/blobio"
	"github.com/jessevdk/go-flags"
)

type PullCommand struct {
	SourcePath       string `required:"true" description:"Location to fetch input blobs from within the bucket."`
	BucketURL string `required:"true" description:"Location of the bucket to fetch blobs from"`
	DestinationPath string `required:"true" description:"Path to inflate with fetched blobs"`
}

func (pc *PullCommand) Execute(args []string) error {
	bucketConfig := blobio.BucketConfig{
		URL:    pc.BucketURL,
	}

	err := blobio.Pull(
		logger,
		context.Background(),
		bucketConfig,
		pc.SourcePath,
		pc.DestinationPath,
	)

	return err

}

var (
	logger lager.Logger
	Pull   PullCommand
)

func main() {
	logger = lager.NewLogger("porter-pull")
	sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
	logger.RegisterSink(sink)

	parser := flags.NewParser(&Pull, flags.HelpFlag|flags.PrintErrors|flags.IgnoreUnknown)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
