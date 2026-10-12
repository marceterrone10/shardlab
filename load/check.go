package load

import "fmt"

// Counts is a row tally for the four CRM tables.
type Counts struct {
	Customers     int64
	Contacts      int64
	Opportunities int64
	Activities    int64
}

// Snapshot is one database after a load.
type Snapshot struct {
	Name     string
	Totals   Counts
	Customer Counts
}

// Check compares the baseline with the shard snapshots. customer is the
// expected presence of one customer and its child rows. expectedShard is the
// only shard that may contain that customer.
func Check(baseline Snapshot, shards []Snapshot, expectedShard string, customer Counts) error {
	if baseline.Name == "" {
		return fmt.Errorf("baseline snapshot has no name")
	}
	if expectedShard == "" {
		return fmt.Errorf("expected shard is required")
	}
	if len(shards) == 0 {
		return fmt.Errorf("at least one shard snapshot is required")
	}

	var (
		found    bool
		sum      Counts
		seenName = map[string]struct{}{}
	)
	for _, shard := range shards {
		if shard.Name == "" {
			return fmt.Errorf("shard snapshot has no name")
		}
		if _, ok := seenName[shard.Name]; ok {
			return fmt.Errorf("duplicate shard snapshot %s", shard.Name)
		}
		seenName[shard.Name] = struct{}{}
		sum = addCounts(sum, shard.Totals)

		if shard.Name == expectedShard {
			found = true
			if shard.Customer != customer {
				return fmt.Errorf("%s customer counts = %+v, want %+v", shard.Name, shard.Customer, customer)
			}
			continue
		}
		if shard.Customer != (Counts{}) {
			return fmt.Errorf("%s contains customer rows %+v", shard.Name, shard.Customer)
		}
	}
	if !found {
		return fmt.Errorf("snapshot for %s is missing", expectedShard)
	}
	if baseline.Customer != customer {
		return fmt.Errorf("baseline customer counts = %+v, want %+v", baseline.Customer, customer)
	}
	if sum != baseline.Totals {
		return fmt.Errorf("shard totals %+v, baseline totals %+v", sum, baseline.Totals)
	}
	return nil
}

func addCounts(left, right Counts) Counts {
	return Counts{
		Customers:     left.Customers + right.Customers,
		Contacts:      left.Contacts + right.Contacts,
		Opportunities: left.Opportunities + right.Opportunities,
		Activities:    left.Activities + right.Activities,
	}
}
