package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
			fmt.Println("Yet another lookout")
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
				newValue, err = check(os.Getenv("ORGANISATION"), os.Getenv("PROJECT"), os.Getenv("CHANNEL"), os.Getenv("FIELD"))
				if err != nil {
					log.Printf("error checking resource, %v", err)
				}
				if newValue != lastValue {
					log.Printf("new value retrieved %v", newValue)
					err = action(os.Getenv("ORGANISATION"), os.Getenv("PROJECT"), os.Getenv("CHANNEL"))
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

func check(organisation, project, channel, field string) (string, error) {
	// TODO: generic target
	// r, err := client.Get(os.Getenv("TARGET_URL"))
	url := fmt.Sprintf("https://quay.io/cnr/api/v1/packages/%s/%s/channels/%s",
		organisation,
		project,
		channel,
	)
	r, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	target := map[string]interface{}{}
	err = json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		return "", err
	}
	if val, ok := target[field]; ok && val != "" {
		return val.(string), nil
	}
	return "", fmt.Errorf("%s field not found in resource", field)
}

func action(organisation, project, channel string) error {
	// TODO trigger generic job
	// cmd := exec.Command("kubectl", "create", "-f /jobs/default.yaml")

	err := switchToTarget(channel)
	if err != nil {
		return err
	}

	releaseExists, err := checkRelease(project)
	if err != nil {
		return err
	}

	cmdName := "helm"
	var cmdArgs []string
	pkg := fmt.Sprintf("quay.io/%s/%s:%s", organisation, project, channel)
	if releaseExists {
		cmdArgs = []string{"registry", "helm", "install", pkg, "-n " + project}

	} else {
		cmdArgs = []string{"registry", "helm", "upgrade", pkg, project}
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func checkRelease(project string) (bool, error) {
	var found bool
	cmd := exec.Command("helm", "list", project)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), project) {
				found = true
				break
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		return false, err
	}

	err = cmd.Wait()
	if err != nil {
		return false, err
	}
	return found, nil
}

func switchToTarget(channel string) error {
	cmd := exec.Command("kubectl", "config", "use-context", "gke_alpha-cluster_europe-west1-b_"+channel)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
