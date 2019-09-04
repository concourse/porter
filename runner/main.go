package main

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"os/exec"
	"time"
)

type RunCommand struct {
	Entrypoint string `required:"true" long:"entrypoint" description:"Command to run on the task image"`
	Args []string `required:"true" long:"args" description:"Arguments for running task image"`
}


func (rc *RunCommand) Execute(args []string) error {

	cmd := exec.Cmd{
		Path:         rc.Entrypoint,
		Args:         rc.Args,
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println("Failed!!!", err)
		// TODO: this err means the task probably failed, and we want to persist this container for hijacking
		return err
	}
	time.Sleep(10 * time.Minute)

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