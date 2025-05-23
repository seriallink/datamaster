package core

import (
	"fmt"
	"time"

	"github.com/seriallink/datamaster/cli/model"

	"github.com/jaswdr/faker"
)

type SimOptions struct {
	RPS       int
	Duration  time.Duration
	Quality   string
	EventType string
	Parallel  int
}

var SupportedScenarios = map[string]func(SimOptions){
	"customer": SimulateCustomerFlow,
	"product":  SimulateProductFlow,
	"purchase": SimulatePurchaseFlow,
	"delivery": SimulateDeliveryFlow,
}

// SimulateCustomerFlow simulates the customer scenario only
func SimulateCustomerFlow(opts SimOptions) {

	err := ensureCDCStarted()
	if err != nil {
		fmt.Println("Failed to start CDC:", err)
		return
	}

	db, err := GetConnection()
	if err != nil {
		fmt.Println("Failed to connect to DB:", err)
		return
	}

	f := faker.New()
	end := time.Now().Add(opts.Duration)
	interval := time.Second / time.Duration(opts.RPS)

	for time.Now().Before(end) {

		customer := model.FakeCustomer(f)

		if shouldCorrupt(f, opts.Quality) {
			customer.CustomerEmail = f.Lorem().Word()
		}

		if err = db.Create(&customer).Error; err != nil {
			fmt.Println("[customer] insert failed:", err)
		}

		time.Sleep(interval)

	}

	fmt.Println("[customer] simulation finished")
}

// SimulatePurchaseFlow simulates the purchase scenario (customer + purchase + delivery)
func SimulatePurchaseFlow(opts SimOptions) {
	fmt.Println("[purchase] Starting simulation...")
	// TODO: implement logic for simulating customers, purchases, and related entities
}

// SimulateProductFlow simulates the product scenario (category + product + review)
func SimulateProductFlow(opts SimOptions) {
	fmt.Println("[product] Starting simulation...")
	// TODO: implement logic for simulating categories and products
}

// SimulateDeliveryFlow simulates the delivery scenario (delivery + tracking)
func SimulateDeliveryFlow(opts SimOptions) {
	fmt.Println("[delivery] Starting simulation...")
	// TODO: implement logic for simulating deliveries and tracking events
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
