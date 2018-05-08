package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Specification struct {
	WatchedFile string `required:"true" split_words:"true"`
	Interval    int    `default:"1"`
	Delay       int    `default:"0"`
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
				fmt.Println(s.WatchedFile, "exists")
				if s.Delay > 0 {
					time.Sleep(time.Duration(s.Delay) * time.Second)
				}
				cmd.Process.Kill()
				break
			}
		}
	}()

	cmd.Wait()
}
