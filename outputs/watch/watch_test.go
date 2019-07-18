package watch_test

import (
	cwatch "github.com/concourse/porter/outputs/watch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
)

var _ = Describe("Watch", func() {

	Context("checkTermination", func() {
		It("returns nil and does nothing when termination is NOT found", func() {
			statuses := []v1.ContainerStatus{
				v1.ContainerStatus{
					Name: "neighboring-container",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "for a test",
						},
					},
				},
				v1.ContainerStatus{
					Name: "task-container",
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{"some-reason", "for a test"},
					},
				},
				v1.ContainerStatus{
					Name: "neighboring-container",
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				},
			}

			onExit0CallCount := 0
			onExitNCallCount := 0

			err := cwatch.CheckTermination(
				"task-container",
				statuses,
				func() error {
					onExit0CallCount++
					return nil
				},
				func(exit int32) error {
					onExitNCallCount++
					return nil
				},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(onExit0CallCount).To(Equal(0))
			Expect(onExitNCallCount).To(Equal(0))
		})

		It("returns nil and invokes onExit0 func when successful termination is found", func() {
			statuses := []v1.ContainerStatus{
				v1.ContainerStatus{
					Name: "neighboring-container",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "for a test",
						},
					},
				},
				v1.ContainerStatus{
					Name: "task-container",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "for a test",
						},
					},
				},
			}

			onExit0CallCount := 0
			onExitNCallCount := 0

			err := cwatch.CheckTermination(
				"task-container",
				statuses,
				func() error {
					onExit0CallCount++
					return nil
				},
				func(exit int32) error {
					onExitNCallCount++
					return nil
				},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(onExit0CallCount).To(Equal(1))
			Expect(onExitNCallCount).To(Equal(0))

		})

		It("returns error and invokes onExitN func when unsuccessful termination is found", func() {
			statuses := []v1.ContainerStatus{
				v1.ContainerStatus{
					Name: "neighboring-container",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "for a test",
						},
					},
				},
				v1.ContainerStatus{
					Name: "task-container",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 2,
							Reason:   "for a test",
						},
					},
				},
			}

			onExit0CallCount := 0
			onExitNCallCount := 0

			err := cwatch.CheckTermination(
				"task-container",
				statuses,
				func() error {
					onExit0CallCount++
					return nil
				},
				func(exit int32) error {
					onExitNCallCount++
					return cwatch.ErrMonitoredContainerExitFail
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cwatch.ErrMonitoredContainerExitFail))
			Expect(onExit0CallCount).To(Equal(0))
			Expect(onExitNCallCount).To(Equal(1))
		})

	})
	/*
		var (
			logger *lagertest.TestLogger
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test")
		})
		Context("waits for watch to finish waiting", func() {

			Context("pushes the output if the termination exitCode is 0", func() {
				fakeClient := fake.NewSimpleClientset()
				watcher := watch.NewFake()
				fakeClient.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))

				pod1 := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test-task-pod",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							v1.ContainerStatus{
								Name: "task-container",
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{"some-reason", "for a test"},
								},
							},
							v1.ContainerStatus{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					},
				}

				// simulate add/update/delete watch events
				watcher.Add(pod1)
				// assert Watch did not terminate

				samePod1ButUpdated := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test-task-pod",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							v1.ContainerStatus{
								Name: "task-container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 0,
										Reason:   "for a test",
									},
								},
							},
							v1.ContainerStatus{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					},
				}
				watcher.Modify(samePod1ButUpdated)
				// assert Watch terminated
			})

			Context("does not push outputs if the termination exitCode is non-zero", func() {
				fakeClient := fake.NewSimpleClientset()
				watcher := watch.NewFake()
				fakeClient.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))

				pod1 := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test-task-pod",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							v1.ContainerStatus{
								Name: "task-container",
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{"some-reason", "for a test"},
								},
							},
							v1.ContainerStatus{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					},
				}
				containerWatcher := &cwatch.ContainerWatcher{
					Client:        fakeClient,
					ContainerName: "task-container",
					PodName:       "test-task-pod",
				}

				// simulate add/update/delete watch events
				watcher.Add(pod1)
				// assert Watch did not terminate

				samePod1ButUpdated := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test-task-pod",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							v1.ContainerStatus{
								Name: "task-container",
								State: v1.ContainerState{
									Terminated: &v1.ContainerStateTerminated{
										ExitCode: 2,
										Reason:   "for a test",
									},
								},
							},
							v1.ContainerStatus{
								Name: "neighboring-container",
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{},
								},
							},
						},
					},
				}
				watcher.Add(samePod1ButUpdated)
				Expect(containerWatcher.Start(logger))
				// assert Watch terminated
			})
		})
	*/
})
