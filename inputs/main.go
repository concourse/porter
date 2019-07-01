package main

import (
	"context"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/porter/k8s"
	"github.com/jessevdk/go-flags"
)

type PullCommand struct {
	SourceURL       string `required:"true" description:"Location to fetch input blobs from"`
	DestinationPath string `required:"true" description:"Path to inflate with fetched blobs"`
}

func (pc *PullCommand) Execute(args []string) error {
	url := os.Getenv("BUCKET_URL")
	bucketConfig := k8s.BucketConfig{
		Type:   *bucketType,
		URL:    url,
		Secret: "notasecret",
	}

	k8s.Pull(
		logger,
		context.Background(),
		bucketConfig,
		pc.SourceURL,
		pc.DestinationPath,
	)

	return nil
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
