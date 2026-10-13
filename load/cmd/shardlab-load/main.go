package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/marceterrone10/shardlab/coordinator"
	"github.com/marceterrone10/shardlab/load"
	"github.com/marceterrone10/shardlab/seed"
)

const (
	defaultBaselineDSN = "postgres://shardlab:shardlab@localhost:5432/shardlab?sslmode=disable"
	defaultShard0DSN   = "postgres://shardlab:shardlab@localhost:5433/shardlab?sslmode=disable"
	defaultShard1DSN   = "postgres://shardlab:shardlab@localhost:5434/shardlab?sslmode=disable"
	defaultShard2DSN   = "postgres://shardlab:shardlab@localhost:5435/shardlab?sslmode=disable"
)

func main() {
	size := flag.String("size", "100k", "number of customers (supports k and m suffixes)")
	randomSeed := flag.Uint64("seed", 42, "deterministic random seed")
	baselineDSN := flag.String("baseline-dsn", defaultBaselineDSN, "baseline PostgreSQL DSN")
	shard0DSN := flag.String("shard0-dsn", defaultShard0DSN, "shard0 PostgreSQL DSN")
	shard1DSN := flag.String("shard1-dsn", defaultShard1DSN, "shard1 PostgreSQL DSN")
	shard2DSN := flag.String("shard2-dsn", defaultShard2DSN, "shard2 PostgreSQL DSN")
	flag.Parse()

	customerCount, err := seed.ParseSize(*size)
	if err != nil {
		exit(err)
	}
	generator, err := seed.NewGenerator(seed.Options{
		Customers: customerCount,
		Seed:      *randomSeed,
	})
	if err != nil {
		exit(err)
	}
	router, err := coordinator.NewRouter([]coordinator.Shard{
		{Name: "shard0"},
		{Name: "shard1"},
		{Name: "shard2"},
	})
	if err != nil {
		exit(err)
	}

	result, err := load.Load(context.Background(), generator, customerCount, router, []load.Target{
		{Name: "baseline", DSN: *baselineDSN, All: true},
		{Name: "shard0", DSN: *shard0DSN},
		{Name: "shard1", DSN: *shard1DSN},
		{Name: "shard2", DSN: *shard2DSN},
	})
	if err != nil {
		exit(err)
	}

	fmt.Printf("loaded %d customers\n", result.Customers)
	for _, snapshot := range result.Snapshots {
		fmt.Printf(
			"%s: %d customers, %d contacts, %d opportunities, %d activities\n",
			snapshot.Name,
			snapshot.Totals.Customers,
			snapshot.Totals.Contacts,
			snapshot.Totals.Opportunities,
			snapshot.Totals.Activities,
		)
	}
	fmt.Printf("verified customer 4 on %s\n", result.ExpectedShard)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
