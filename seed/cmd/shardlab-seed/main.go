package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/marceterrone10/shardlab/seed"
)

func main() {
	size := flag.String("size", "100k", "number of customers (supports k and m suffixes)")
	randomSeed := flag.Uint64("seed", 42, "deterministic random seed")
	output := flag.String("output", ".data/seed", "output directory")
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

	counts, err := seed.WriteCSV(*output, generator)
	if err != nil {
		exit(err)
	}

	fmt.Printf(
		"generated %d customers, %d contacts, %d opportunities and %d activities in %s\n",
		counts.Customers,
		counts.Contacts,
		counts.Opportunities,
		counts.Activities,
		*output,
	)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
