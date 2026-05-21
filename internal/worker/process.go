package worker

import (
	"time"
	"errors"
	"math/rand"

	"github.com/sanin7k/conveyor/internal/job"
)

func process(job job.Job) error {
	switch job.Type {
	case "sendEmail":
		time.Sleep(2 * time.Second)
		if rand.Float32() < 0.2 {
			return errors.New("email service unavailable")
		}
		return nil
	
	case "resizeImage":
		time.Sleep(3 * time.Second)
		if rand.Float32() < 0.15 {
			return errors.New("image resize failed")
		}
		return nil
	case "generateReport":
		time.Sleep(5 * time.Second)
		if rand.Float32() < 0.10 {
			return errors.New("report generation failed")
		}
		return nil
	default:
		return errors.New("unknown job type")
	}
}

