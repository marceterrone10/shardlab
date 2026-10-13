package load

import (
	"context"
	"testing"

	"github.com/marceterrone10/shardlab/seed"
)

func TestLoadRejectsFewerThanFourCustomers(t *testing.T) {
	generator, err := seed.NewGenerator(seed.Options{Customers: 3, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	_, err = Load(context.Background(), generator, 3, testRouter(t), nil)
	if err == nil {
		t.Fatal("Load() returned no error")
	}
}

func TestLoadRejectsMissingShardTarget(t *testing.T) {
	generator, err := seed.NewGenerator(seed.Options{Customers: 4, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	_, err = Load(context.Background(), generator, 4, testRouter(t), []Target{
		{Name: "baseline", DSN: "postgres://shardlab@localhost/shardlab", All: true},
	})
	if err == nil {
		t.Fatal("Load() returned no error")
	}
}
