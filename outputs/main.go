package main

import (
	"os"

	"code.cloudfoundry.org/lager"
	cwatch "github.com/concourse/porter/outputs/watch"
	"github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type PushCommand struct {
	PodName       string `required:"true" positional-args:"yes" description:"Pod to watch"`
	ContainerName string `required:"true" positional-args:"yes" description:"Container to wait till completion"`

	SourcePath     string `required:"true" description:"Path to outputs dir intended to be pushed"`
	DestinationURL string `required:"true" description:"Location inside provided bucket to deposit output blobs"`
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

	return watcher.Start(logger)
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
