package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/marceterrone10/shardlab/analyzer"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("shardlab-analyzer", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	minFrequency := flags.Int64("min-frequency", 1, "minimum aggregate join frequency")
	inputPath := flags.String("input", "", "JSON input file; defaults to stdin")
	if err := flags.Parse(args); err != nil {
		return err
	}

	input := stdin
	if *inputPath != "" {
		file, err := os.Open(*inputPath)
		if err != nil {
			return fmt.Errorf("open observations: %w", err)
		}
		defer file.Close()
		input = file
	}

	var observations []analyzer.JoinObservation
	if err := json.NewDecoder(input).Decode(&observations); err != nil {
		return fmt.Errorf("decode observations: %w", err)
	}

	suggestions, err := analyzer.Suggest(observations, *minFrequency)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(suggestions)
}
