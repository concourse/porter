package main

import (
	//"context"
	"errors"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
	//"github.com/concourse/porter/k8s"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var ErrMonitoredContainerExitFail = errors.New("monitored container did not exit well")

type PushCommand struct {
	PodName       string `required:"true" positional-args:"yes" description:"Pod to watch"`
	ContainerName string `required:"true" positional-args:"yes" description:"Container to wait till completion"`

	SourcePath     string `required:"true" description:"Path to outputs dir intended to be pushed"`
	DestinationURL string `required:"true" description:"Location inside provided bucket to deposite output blobs"`
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

	watch, err := clientset.CoreV1().Pods("default").Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", pc.PodName),
	})
	if err != nil {
		logger.Error("failed to find pod", err)
		return err
	}

	for event := range watch.ResultChan() {
		// We only care when event.Type == MODIFIED
		// MODIFIED events occur when ContainerStatus is
		// updated, which contains exit code

		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			logger.Debug("failed to typecast, this should never happen...")
			continue
		}

		err := checkTermination(
			pc.ContainerName,
			pod.Status.ContainerStatuses,
			func() error {
				logger.Info("monitored container exited successfully", lager.Data{})
				// *push outputs here*
				// k8s.Push(logger, context.Background(), bucketConfig, *sourceKey, *destionationPath)
				return nil
			},
			func(exitCode int32) error {
				logger.Info("monitored container returned non-zero exit code", lager.Data{
					"podname":  pc.PodName,
					"ExitCode": exitCode,
				})
				return ErrMonitoredContainerExitFail
			},
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func checkTermination(
	container string,
	statuses []v1.ContainerStatus,
	onExit0 func() error,
	onExitN func(exitCode int32) error,
) error {
	for _, containerInfo := range statuses {
		state := containerInfo.State

		if containerInfo.Name == container && state.Terminated != nil {
			if state.Terminated.ExitCode == 0 {
				return onExit0()
			} else if state.Terminated.ExitCode != 0 {
				return onExitN(state.Terminated.ExitCode)
			}
		}
	}
	return nil
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
