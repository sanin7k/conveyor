package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sanin7k/conveyor/internal/job"
)

var spinnerFrame = []rune{'|', '/', '-', '\\'}
const doneFrame rune = '✓'
const failedFrame rune = '✗'

func main() {
	counter := 0

	var jobs []job.Job
	for {
		jobs = fetchJobs()
		clearScreen()
		printHeader(len(jobs))
		for _, job := range jobs {
			printRow(job, counter)
		}
		counter++
		time.Sleep(time.Second)
	}

}

func fetchJobs() []job.Job {
	resp, err := http.Get("http://localhost:8080/jobs")
	if (err != nil) {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}

	var jobs []job.Job

	err = json.NewDecoder(resp.Body).Decode(&jobs)
	if err != nil {
		log.Println(err)
		return nil
	}

	return jobs
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printHeader(numJobs int) {
	fmt.Println("Conveyor - ", numJobs, " jobs")
	fmt.Println()

	fmt.Print(strings.Repeat(" ", 5))
	fmt.Print("  ID        ")
	fmt.Print("  Job Type        ")
	fmt.Print("  Status  ")
	fmt.Print("  Attempts  ")

	fmt.Println()
	fmt.Println()
}


func printRow(j job.Job, counter int) {
	var frame rune
	switch j.Status {
	case job.Done:
		frame = doneFrame
	case job.Failed:
		frame = failedFrame
	default:
		frame = spinnerFrame[counter % 4]
	}

	fmt.Printf(" [%c] ", frame)
	fmt.Print("  ", j.ID[:8], "  ")

	jobType := j.Type
	if len(jobType) > 16 {
		jobType = jobType[:12] + ".."
	}
	padding := 16 - len(jobType)
	fmt.Print("  ", jobType, strings.Repeat(" ", padding))

	padding = 8 - len(string(j.Status))
	fmt.Print("  ", j.Status, strings.Repeat(" ", padding))

	fmt.Print("  ", j.Attempts, "/3")

	fmt.Println()
}

