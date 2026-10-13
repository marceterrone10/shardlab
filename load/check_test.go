package load

import "testing"

func TestCheckAcceptsMatchingDistribution(t *testing.T) {
	customer := Counts{Customers: 1, Contacts: 1, Opportunities: 1, Activities: 2}
	baseline := Snapshot{
		Name:     "baseline",
		Totals:   Counts{Customers: 4, Contacts: 8, Opportunities: 8, Activities: 12},
		Customer: customer,
	}
	shards := []Snapshot{
		{Name: "shard0", Totals: Counts{Customers: 2, Contacts: 4, Opportunities: 4, Activities: 6}, Customer: customer},
		{Name: "shard1", Totals: Counts{Customers: 1, Contacts: 2, Opportunities: 2, Activities: 3}},
		{Name: "shard2", Totals: Counts{Customers: 1, Contacts: 2, Opportunities: 2, Activities: 3}},
	}

	if err := Check(baseline, shards, "shard0", customer); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckRejectsCustomerOnAnotherShard(t *testing.T) {
	customer := Counts{Customers: 1, Contacts: 1, Opportunities: 1, Activities: 1}
	baseline := Snapshot{Name: "baseline", Totals: Counts{Customers: 1}, Customer: customer}
	shards := []Snapshot{
		{Name: "shard0", Totals: Counts{Customers: 1}},
		{Name: "shard1", Totals: Counts{}, Customer: customer},
	}

	if err := Check(baseline, shards, "shard0", customer); err == nil {
		t.Fatal("Check() returned no error")
	}
}

func TestCheckRejectsMismatchedTotals(t *testing.T) {
	customer := Counts{Customers: 1}
	baseline := Snapshot{Name: "baseline", Totals: Counts{Customers: 4}, Customer: customer}
	shards := []Snapshot{
		{Name: "shard0", Totals: Counts{Customers: 1}, Customer: customer},
	}

	if err := Check(baseline, shards, "shard0", customer); err == nil {
		t.Fatal("Check() returned no error")
	}
}
