package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type logresp struct {
	Pos  int    `json:"pos"`
	Out  string `json:"out"`
	Time int    `json:"time"`
}

type stepsresp struct {
	Number int    `json:"number"`
	Status string `json:"status"`
}

type stagesresp struct {
	Status string      `json:"status"`
	Name   string      `json:"name"`
	Number int         `json:"number"`
	Steps  []stepsresp `json:"steps"`
}

type buildresp struct {
	Status string       `json:"status"`
	Stages []stagesresp `json:"stages"`
}

func main() {
	DRONE_PULL_REQUEST := os.Getenv("DRONE_PULL_REQUEST")
	DRONE_REPO_OWNER := os.Getenv("DRONE_REPO_OWNER")
	DRONE_REPO_NAME := os.Getenv("DRONE_REPO_NAME")
	DRONE_ACCESS_TOKEN := os.Getenv("DRONE_ACCESS_TOKEN")
	DRONE_HOST := os.Getenv("DRONE_HOST")
	DRONE_BUILD_NUMBER := os.Getenv("DRONE_BUILD_NUMBER")
	GITHUB_ACCESS_TOKEN := os.Getenv("GITHUB_ACCESS_TOKEN")

	var drone_msg string

	url := "https://" + DRONE_HOST + "/api/repos/" + DRONE_REPO_OWNER + "/" + DRONE_REPO_NAME + "/builds/" + DRONE_BUILD_NUMBER

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}

	req.Header.Add("Authorization", "Bearer "+DRONE_ACCESS_TOKEN)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}

	var myStoredVariable buildresp
	var stageNumber int = 0
	var stepNumber int = 0
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}

	err = json.Unmarshal(body, &myStoredVariable)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}

	if myStoredVariable.Status == "success" {
		fmt.Print("Build Success")
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	for _, stage := range myStoredVariable.Stages {
		if stage.Status == "failure" {
			for _, step := range stage.Steps {
				if step.Status == "failure" {
					stageNumber = stage.Number
					stepNumber = step.Number
				}
			}
		}
	}
	if stageNumber == 0 || stepNumber == 0 {
		fmt.Println("No failures found")
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	urlLog := "https://" + DRONE_HOST + "/api/repos/" + DRONE_REPO_OWNER + "/" + DRONE_REPO_NAME + "/builds/" + DRONE_BUILD_NUMBER + "/logs/" + strconv.Itoa(stageNumber) + "/" + strconv.Itoa(stepNumber)
	reqLog, err := http.NewRequest("GET", urlLog, nil)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	reqLog.Header.Add("Authorization", "Bearer "+DRONE_ACCESS_TOKEN)

	resLog, err := http.DefaultClient.Do(reqLog)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	defer resLog.Body.Close()
	bodyLog, err := ioutil.ReadAll(resLog.Body)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	var logs []logresp
	err = json.Unmarshal(bodyLog, &logs)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	for _, log := range logs {
		drone_msg += log.Out + "<br>"
	}
	fmt.Println(drone_msg)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_ACCESS_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	comment := github.IssueComment{
		Body: &drone_msg,
	}
	drone_pr_no, err := strconv.Atoi(DRONE_PULL_REQUEST)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	_, _, err = client.Issues.CreateComment(ctx, DRONE_REPO_OWNER, DRONE_REPO_NAME, drone_pr_no, &comment)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Exiting Gracefully")
		os.Exit(0)
	}
	fmt.Print("Comment added to PR")
}
