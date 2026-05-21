package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sanin7k/conveyor/internal/dispatcher"
	"github.com/sanin7k/conveyor/internal/job"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	d := dispatcher.New(ctx, 10, 4)
	d.Start()

	types := []string{
		"sendEmail",
		"generateReport",
		"resizeImage",
	}
	
	for i := range 10 {
		j := job.Job{
			ID: strconv.Itoa(i),
			Type: types[rand.Intn(len(types))],
			Payload: nil,
			Status: "pending",
			Attempts: 0,
		}
		d.Submit(j)
		fmt.Println("Submitted Job: [", j.ID, "] ", j.Type)
	}

	<-ctx.Done()

	fmt.Println("Signal received")
	fmt.Println("Shutting down")

	d.Shutdown()
}

