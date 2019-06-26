package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"context"
	"../k8s"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"code.cloudfoundry.org/lager"
	//"github.com/concourse/concourse/atc/k8s"

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

	fmt.Println("running your watcher cmd with args", *podname, *containername)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	max_retries := 20
	for i := 1; i <= max_retries; i++ {
		pod, err := clientset.CoreV1().Pods("default").Get(*podname, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, c := range pod.Status.ContainerStatuses {
			fmt.Println("searching containers", c.Name)
			if c.Name == *containername {
				fmt.Println("looping on containers", c.Name)
				if c.State.Terminated != nil && c.State.Terminated.ExitCode == 0 {
					fmt.Println("ready to upload output")
					url := os.Getenv("BUCKET_URL")
					bucketConfig := k8s.BucketConfig{
						Type: *bucketType,
						URL: url,
						Secret: "notasecret",
					}
					k8s.Push(lager.NewLogger("output_pusher"), context.Background(), bucketConfig, *sourceKey, *destionationPath)
					os.Exit(0)
				} else if c.State.Terminated != nil && c.State.Terminated.ExitCode != 0 {
					fmt.Println("container step failed with exit code", c.State.Terminated.ExitCode)
					os.Exit(5)
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

