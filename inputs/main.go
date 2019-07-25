package main

import (
	"context"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/porter/blobio"
	"github.com/jessevdk/go-flags"
)

type PullCommand struct {
	SourcePath       string `required:"true" long:"source-path" description:"Location to fetch input blobs from within the bucket."`
	BucketURL string `required:"true" long:"bucket-url" description:"Location of the bucket to fetch blobs from"`
	DestinationPath string `required:"true" long:"destination-path" description:"Path to inflate with fetched blobs"`
}

func (pc *PullCommand) Execute(args []string) error {
	bucketConfig := blobio.BucketConfig{
		URL:    pc.BucketURL,
	}

	return blobio.Pull(
		logger,
		context.Background(),
		bucketConfig,
		pc.SourcePath,
		pc.DestinationPath,
	)

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

	err = Pull.Execute(os.Args)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
