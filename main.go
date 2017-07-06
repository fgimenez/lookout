package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Println("0.1.0")
			return
		}

		if os.Args[1] == "--help" {
			fmt.Println("Yet another tacoshop")
			return
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int)
	tick := time.Tick(10 * time.Second)

	log.Println("Start polling...")
	var lastValue, newValue string
	var err error
	go func() {
		for {
			select {
			case <-tick:
				log.Println("tick")
				newValue, err = check()
				if err != nil {
					log.Printf("error checking resource, %v", err)
				}
				if newValue != lastValue {
					log.Printf("new value retrieved %v", newValue)
					err = action()
					if err != nil {
						log.Printf("error executing action, %v", err)
					} else {
						lastValue = newValue
					}
				}
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					log.Println("Stopping")
					exitChan <- 0
				default:
					log.Println("Unknown signal")
					exitChan <- 1
				}
			}
		}
	}()
	code := <-exitChan
	os.Exit(code)
}

func check() (string, error) {
	r, err := client.Get(os.Getenv("TARGET_URL"))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	target := map[string]interface{}{}
	err = json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		return "", err
	}
	if val, ok := target[os.Getenv("FIELD")]; ok && val != "" {
		return val.(string), nil
	}
	return "", fmt.Errorf("%s field not found in resource", os.Getenv("FIELD"))
}

func action() error {
	cmd := exec.Command("kubectl", "create", "-f /jobs/default.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
