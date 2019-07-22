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

var ErrMonitoredContainerExitFail = errors.New("monitored container exited with error")
var ErrContainerNotFoundInPod = errors.New("container doesn't exist under specified Pod")
var ErrWatcherclosedWithoutTerminationStatus = errors.New("Watcher result channel closed without container being terminated")

type ContainerWatcher struct {
	Client        kubernetes.Interface
	ContainerName string
	PodName       string
}

func (cw *ContainerWatcher) Wait(logger lager.Logger) error {
	watch, err := cw.Client.CoreV1().Pods("default").Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", cw.PodName),
	})
	defer watch.Stop()

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

		if isContainerInPod(pod, cw) != true {
			return ErrContainerNotFoundInPod
		}

		for _, containerInfo := range pod.Status.ContainerStatuses {
			state := containerInfo.State

			if containerInfo.Name == cw.ContainerName && state.Terminated != nil {
				if state.Terminated.ExitCode == 0 {
					logger.Info("monitored container exited successfully", lager.Data{})
					return nil
				} else if state.Terminated.ExitCode != 0 {
					logger.Info("monitored container returned non-zero exit code", lager.Data{
						"podname":  cw.PodName,
						"ExitCode": state.Terminated.ExitCode,
					})
					return ErrMonitoredContainerExitFail
				}
			}
		}
	}
	return ErrWatcherclosedWithoutTerminationStatus
}

func isContainerInPod(pod *v1.Pod, cw *ContainerWatcher) bool {
	found := false
	for _, container := range pod.Spec.Containers {
		if container.Name == cw.ContainerName {
			found = true
		}
	}
	return found
}


