package seed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := map[string]int64{
		"10":   10,
		"100k": 100_000,
		"1M":   1_000_000,
	}

	for input, want := range tests {
		got, err := ParseSize(input)
		if err != nil {
			t.Fatalf("ParseSize(%q) error = %v", input, err)
		}
		if got != want {
			t.Errorf("ParseSize(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestGeneratorIsDeterministic(t *testing.T) {
	generator, err := NewGenerator(Options{Customers: 10, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	first, err := generator.Customer(5)
	if err != nil {
		t.Fatalf("Customer(5) error = %v", err)
	}
	second, err := generator.Customer(5)
	if err != nil {
		t.Fatalf("Customer(5) error = %v", err)
	}

	if first.Customer != second.Customer || len(first.Activities) != len(second.Activities) {
		t.Fatalf("Customer(5) returned different values: %#v and %#v", first, second)
	}
}

func TestWriteCSVCreatesAllTables(t *testing.T) {
	generator, err := NewGenerator(Options{Customers: 2, Seed: 42})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	outputDirectory := t.TempDir()
	counts, err := WriteCSV(outputDirectory, generator)
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}
	if counts.Customers != 2 {
		t.Fatalf("WriteCSV() customer count = %d, want 2", counts.Customers)
	}

	for _, name := range []string{
		"clientes.csv",
		"contactos.csv",
		"oportunidades.csv",
		"actividades.csv",
	} {
		info, err := os.Stat(filepath.Join(outputDirectory, name))
		if err != nil {
			t.Errorf("%s was not created: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", name)
		}
	}
}
