package load

import (
	"testing"

	"github.com/marceterrone10/shardlab/coordinator"
	"github.com/marceterrone10/shardlab/seed"
)

func TestDestinationAssignsCustomersToShards(t *testing.T) {
	router := testRouter(t)
	generator, err := seed.NewGenerator(seed.Options{Customers: 4, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	wants := map[int64]string{
		1: "shard0",
		2: "shard1",
		3: "shard2",
		4: "shard0",
	}
	for customerID, want := range wants {
		bundle, err := generator.Customer(customerID)
		if err != nil {
			t.Fatalf("Customer(%d) error = %v", customerID, err)
		}
		got, err := Destination(bundle, router)
		if err != nil {
			t.Fatalf("Destination(%d) error = %v", customerID, err)
		}
		if got != want {
			t.Errorf("Destination(%d) = %q, want %q", customerID, got, want)
		}
	}
}

func TestDestinationKeepsContactWithCustomer(t *testing.T) {
	router := testRouter(t)
	generator, err := seed.NewGenerator(seed.Options{Customers: 4, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	bundle, err := generator.Customer(4)
	if err != nil {
		t.Fatalf("Customer(4) error = %v", err)
	}

	var contact41 bool
	for _, contact := range bundle.Contacts {
		if contact.ID == 41 && contact.CustomerID == 4 {
			contact41 = true
		}
	}
	if !contact41 {
		t.Fatal("customer 4 bundle does not contain contact 41")
	}

	got, err := Destination(bundle, router)
	if err != nil {
		t.Fatalf("Destination() error = %v", err)
	}
	// Contact 41 would land on shard1 if its own ID were the partition key:
	// (41 - 1) % 3 = 1. The bundle follows customer 4 instead.
	if got != "shard0" {
		t.Fatalf("Destination() = %q, want shard0", got)
	}
}

func TestDestinationRejectsChildFromAnotherCustomer(t *testing.T) {
	router := testRouter(t)
	bundle := seed.CustomerBundle{
		Customer: seed.Customer{ID: 4},
		Contacts: []seed.Contact{{ID: 41, CustomerID: 5}},
	}

	if _, err := Destination(bundle, router); err == nil {
		t.Fatal("Destination() returned no error")
	}
}

func testRouter(t *testing.T) *coordinator.Router {
	t.Helper()
	router, err := coordinator.NewRouter([]coordinator.Shard{
		{Name: "shard0"},
		{Name: "shard1"},
		{Name: "shard2"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	return router
}
