package watch_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/concourse/porter/outputs/watch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var (
	logger *lagertest.TestLogger
)

var _ = Describe("Watch", func() {
	var (
	fakeClient *fake.Clientset
	watcher *watch.FakeWatcher
	pod *v1.Pod
	updatedPod *v1.Pod
	containerWatcher *ContainerWatcher
	)

	BeforeEach(func(){
		logger = lagertest.NewTestLogger("test")
		fakeClient = fake.NewSimpleClientset()
		watcher = watch.NewFake()

		fakeClient.PrependWatchReactor("pods", testing.DefaultWatchReactor(watcher, nil))

		pod = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-task-pod",
			},
			Spec: v1.PodSpec{
				NodeName: "node1",
				Containers: []v1.Container{
					{
						Name: "task-container",
					},
					{
						Name: "neighboring-container",
					},
				},
			},

			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "task-container",
						State: v1.ContainerState{
							Waiting: &v1.ContainerStateWaiting{"some-reason", "for a test"},
						},
					},
					{
						Name: "neighboring-container",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
		}
	})


		Context("when Container.State is TERMINATED", func() {
			BeforeEach(func() {

				updatedPod = &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test-task-pod",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
						Containers: []v1.Container{
							{
								Name: "task-container",
							},
							{
								Name: "neighboring-container",
							},
						},
					},
				}

				containerWatcher = &ContainerWatcher{
					Client:        fakeClient,
					ContainerName: "task-container",
					PodName:       "test-task-pod",
				}
			})
			XIt("and exitCode is Zero, Watch returns nil error", func() {
				go func () {
					watcher.Add(pod)

					updatedPodStatus := v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "task-container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
										Reason:   "for a test",
									},
								},
							},
							{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					}
					updatedPod.Status = updatedPodStatus

					watcher.Modify(updatedPod)

				}()

				err := containerWatcher.Wait(logger)
				Expect(err).ToNot(HaveOccurred())
			})

			XIt("exitCode is NOT Zero, Watch returns an error", func() {

				go func () {
					watcher.Add(pod)

					updatedPod.Status = v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: "task-container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 2,
										Reason:   "for a test",
									},
								},
							},
							{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					}

					watcher.Add(updatedPod)
				}()


				err := containerWatcher.Wait(logger)
				Expect(err).To(Equal(ErrMonitoredContainerExitFail))
			})
		})

		Context("when Container name is not found in the Pod", func() {
		BeforeEach(func() {
			containerWatcher = &ContainerWatcher{
				Client:        fakeClient,
				ContainerName: "some-non-existent-task-container",
				PodName:       "test-task-pod",
			}
		})
		XIt("Watch returns an error", func() {
			go func () {
				watcher.Add(pod)
			}()

			err := containerWatcher.Wait(logger)
			Expect(err).To(Equal(ErrContainerNotFoundInPod))
		})


	})

})
