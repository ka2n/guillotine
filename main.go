package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Specification struct {
	WatchedFile string `required:"true" split_words:"true"`
	Interval    int    `default:"1"`
	Delay       int    `default:"0"`
	KillDelay   int    `default:"30"`
}

func main() {
	var s Specification

	if err := envconfig.Process("guillotine", &s); err != nil {
		log.Fatal(err.Error())
		return
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: cmd args...")
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	fmt.Println("Watch File:", s.WatchedFile)
	if err := cmd.Start(); err != nil {
		return
	}

	go func() {
		for {
			time.Sleep(time.Duration(s.Interval) * time.Second)
			if _, err := os.Stat(s.WatchedFile); err == nil {
				if s.Delay > 0 {
					duration := time.Duration(s.Delay) * time.Second
					fmt.Printf("%s exists, send SIGTERM after %v\n", s.WatchedFile, duration)
					time.Sleep(duration)
				} else {
					fmt.Println(s.WatchedFile, "exists, send SIGTERM immediately")
				}
				cmd.Process.Signal(syscall.SIGTERM)
				time.Sleep(time.Duration(s.KillDelay) * time.Second)
				fmt.Println("KillDelay passed, send SIGKILL")
				cmd.Process.Kill()
				break
			}
		}
	}()

	cmd.Wait()

	os.Exit(getExitStatus(cmd))
}

func getExitStatus(cmd *exec.Cmd) int {
	if cmd.ProcessState == nil {
		return 1
	}

	sys := cmd.ProcessState.Sys()
	if sys != nil {
		if code, ok := sys.(syscall.WaitStatus); ok {
			return code.ExitStatus()
		}
	}

	if !cmd.ProcessState.Success() {
		return 1
	}

	return 0
}
