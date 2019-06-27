package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/porter/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {

	podname := flag.String("pod", "", "name of the pod")
	containername := flag.String("container", "", "name of the container to wait for")
	bucketType := flag.String("bucketType", "", "bucket type, allowed values are s3 and gcs")
	sourceKey := flag.String("sourceKey", "", "source to be downloaded")
	destionationPath := flag.String("destionationPath", "", "location to place the downloaded artifact")

	flag.Parse()

	logger := lager.NewLogger("output-push")
	logger.Info("")
	fmt.Println("running your watcher cmd with args", *podname, *containername)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("failed to retrieve cluster config", err)
		os.Exit(1)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("failed to create client with fetched config", err)
		os.Exit(1)
	}

	// another approach? seems more native with the Watch() client endpoint
	watch, err := clientset.CoreV1().Pods("default").Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", *podname),
	})
	if err != nil {
		logger.Error("failed to find pod", err)
		os.Exit(1)
	}

	go func() {
		for event := range watch.ResultChan() {
			// we only care when event.Type == MODIFIED, This event
			// occurs when ContainerStatus is updated
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}

			for _, containerInfo := range pod.Status.ContainerStatuses {
				terminationInfo := containerInfo.State.Terminated

				if containerInfo.Name == *containername && terminationInfo.Terminated != nil {

					if terminationInfo.ExitCode == 0 {
						url := os.Getenv("BUCKET_URL")
						bucketConfig := k8s.BucketConfig{
							Type:   *bucketType,
							URL:    url,
							Secret: "notasecret",
						}

						k8s.Push(logger, context.Background(), bucketConfig, *sourceKey, *destionationPath)

						os.Exit(0)
					} else if terminationInfo.ExitCode != 0 {
						logger.Info("output producer container returned non-zero exit code")
						os.Exit(1)
					}
				}
			}
		}
	}()

	// simple approach and works.
	for {
		pod, err := clientset.CoreV1().Pods("default").Get(*podname, metav1.GetOptions{})
		if err != nil {
			logger.Error("failed to find pod", err)
			os.Exit(1)
		}

		for _, c := range pod.Status.ContainerStatuses {
			if c.Name == *containername && c.State.Terminated != nil {
				if c.State.Terminated.ExitCode == 0 {
					url := os.Getenv("BUCKET_URL")
					bucketConfig := k8s.BucketConfig{
						Type:   *bucketType,
						URL:    url,
						Secret: "notasecret",
					}

					k8s.Push(logger, context.Background(), bucketConfig, *sourceKey, *destionationPath)

					os.Exit(0)
				} else if c.State.Terminated.ExitCode != 0 {
					logger.Info("output producer container returned non-zero exit code")
					os.Exit(1)
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}
