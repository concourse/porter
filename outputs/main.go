package main

import (
	"context"
	"github.com/concourse/porter/blobio"
	"os"

	"code.cloudfoundry.org/lager"
	cwatch "github.com/concourse/porter/outputs/watch"
	"github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PushCommand struct {
	PodName       string `required:"true" positional-args:"yes" description:"Pod to watch"`
	ContainerName string `required:"true" positional-args:"yes" description:"Container to wait till completion"`

	SourcePath      string `required:"true" description:"Location to fetch input blobs from within the bucket."`
	BucketURL       string `required:"true" description:"Location of the bucket to fetch blobs from"`
	DestinationPath string `required:"true" description:"Path to inflate with fetched blobs"`
}

func (pc *PushCommand) Execute(args []string) error {
	logger.Debug("push-execute", lager.Data{
		"podname":       pc.PodName,
		"containername": pc.ContainerName,
	})

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("failed to retrieve cluster config", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("failed to create client with fetched config", err)
		return err
	}

	watcher := cwatch.ContainerWatcher{
		Client:        clientset,
		ContainerName: pc.ContainerName,
		PodName:       pc.PodName,
	}

	err = watcher.Wait(logger)
	if err != nil {
		return err
	}

	bucketConfig := blobio.BucketConfig{
		URL: pc.BucketURL,
	}

	return  blobio.Push(
		logger,
		context.Background(),
		bucketConfig,
		pc.SourcePath,
		pc.DestinationPath,
	)

}

var (
	logger lager.Logger
	Push   PushCommand
)

func main() {
	logger = lager.NewLogger("porter-push")
	sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
	logger.RegisterSink(sink)

	parser := flags.NewParser(&Push, flags.HelpFlag|flags.PrintErrors|flags.IgnoreUnknown)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
