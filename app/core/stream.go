package core

import (
	"fmt"
	"time"

	"github.com/jaswdr/faker"
)

type SimOptions struct {
	RPS      int
	Duration time.Duration
	Quality  string
	Parallel int
}

var SupportedStreams = map[string]func(SimOptions){
	"profile": SimulateProfileStream,
	"review":  SimulateReviewStream,
}

// SimulateProfileStream simulates the customer scenario only
func SimulateProfileStream(opts SimOptions) {

	//err := ensureCDCStarted()
	//if err != nil {
	//	fmt.Println("Failed to start CDC:", err)
	//	return
	//}
	//
	//db, err := GetConnection()
	//if err != nil {
	//	fmt.Println("Failed to connect to DB:", err)
	//	return
	//}
	//
	//f := faker.New()
	//end := time.Now().Add(opts.Duration)
	//interval := time.Second / time.Duration(opts.RPS)
	//
	//for time.Now().Before(end) {
	//
	//	profile := model.FakeProfile(f)
	//
	//	if shouldCorrupt(f, opts.Quality) {
	//		profile.Email = f.Lorem().Word()
	//	}
	//
	//	if err = db.Create(&profile).Error; err != nil {
	//		fmt.Println("[profile] insert failed:", err)
	//	}
	//
	//	time.Sleep(interval)
	//
	//}
	//
	//fmt.Println("[customer] simulation finished")
}

// SimulateReviewStream simulates the purchase scenario (customer + purchase + delivery)
func SimulateReviewStream(opts SimOptions) {
	fmt.Println("[review] Starting simulation...")
	// TODO: implement logic for simulating
}

func shouldCorrupt(f faker.Faker, quality string) bool {
	switch quality {
	case "high":
		return false
	case "medium":
		return f.IntBetween(0, 9) < 2 // ~20%
	case "low":
		return f.IntBetween(0, 1) == 0 // ~50%
	default:
		return false
	}
}
