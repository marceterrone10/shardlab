package coordinator

import (
	"fmt"
	"os"
	"strconv"
)

const defaultShardCount = 3

// Config contains the coordinator process configuration.
type Config struct {
	HTTPAddress string
	Shards      []Shard
}

// LoadConfig reads coordinator configuration from the environment.
func LoadConfig() (Config, error) {
	count, err := envInt("SHARDLAB_SHARD_COUNT", defaultShardCount)
	if err != nil {
		return Config{}, err
	}
	if count < 1 {
		return Config{}, fmt.Errorf("SHARDLAB_SHARD_COUNT must be greater than zero")
	}

	shards := make([]Shard, 0, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("shard%d", i)
		defaultDSN := fmt.Sprintf(
			"postgres://shardlab:shardlab@%s:5432/shardlab?sslmode=disable",
			name,
		)
		shards = append(shards, Shard{
			Name: name,
			DSN:  envString(fmt.Sprintf("SHARDLAB_SHARD_%d_DSN", i), defaultDSN),
		})
	}

	return Config{
		HTTPAddress: envString("SHARDLAB_HTTP_ADDRESS", ":8080"),
		Shards:      shards,
	}, nil
}

func envString(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok && value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) (int, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}
