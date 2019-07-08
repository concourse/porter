package watch

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var ErrMonitoredContainerExitFail = errors.New("monitored container did not exit well")

type ContainerWatcher struct {
	Client        kubernetes.Interface
	ContainerName string
	PodName       string
}

func (cw *ContainerWatcher) Start(logger lager.Logger) error {
	watch, err := cw.Client.CoreV1().Pods("default").Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", cw.PodName),
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

		err := CheckTermination(
			cw.ContainerName,
			pod.Status.ContainerStatuses,
			func() error {
				logger.Info("monitored container exited successfully", lager.Data{})
				// *push outputs here*
				return nil
			},
			func(exitCode int32) error {
				logger.Info("monitored container returned non-zero exit code", lager.Data{
					"podname":  cw.PodName,
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

func CheckTermination(
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
