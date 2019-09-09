package main

import (
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type RunCommand struct {
	Entrypoint string `required:"true" long:"entrypoint" description:"Command to run on the task image"`
	Args string `required:"true" long:"args" description:"Arguments for running task image"`
	ConfigMapName string `required:"true" long:"configmap" description:"ConfigMap to update task status"`
}


func (rc *RunCommand) Execute(args []string) error {
	var exitCode int
	var exitErrMsg string


	cmd := exec.Command(rc.Entrypoint, "-c", rc.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
				exitErrMsg = exiterr.Error()
			}
		}
	}

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

	cm, err := clientset.CoreV1().ConfigMaps("default").Get(rc.ConfigMapName, metav1.GetOptions{})
	if err != nil {
		logger.Error("failed get config map", err)
		return err
	}
	cm.Data["exitcode"] = strconv.Itoa(exitCode)
	cm.Data["err"] = exitErrMsg

	_, err = clientset.CoreV1().ConfigMaps("default").Update(cm)
	if err != nil {
		logger.Error("failed to update config map", err)
		return err
	}

	if exitCode != 0 {
		time.Sleep(10 * time.Minute)
	}

	return nil
}


var (
	logger lager.Logger
	Runner RunCommand
)

func main() {
	logger = lager.NewLogger("run-task-command")
	sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
	logger.RegisterSink(sink)

	parser := flags.NewParser(&Runner, flags.HelpFlag|flags.PrintErrors|flags.IgnoreUnknown)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	err = Runner.Execute(os.Args)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}