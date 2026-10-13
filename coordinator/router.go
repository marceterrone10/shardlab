package coordinator

import "fmt"

// Shard identifies one PostgreSQL node. DSN is kept here so the query
// executor can be introduced without changing the routing API.
type Shard struct {
	Name string `json:"name"`
	DSN  string `json:"-"`
}

// Router maps customer-scoped queries to a shard.
type Router struct {
	shards []Shard
}

func NewRouter(shards []Shard) (*Router, error) {
	if len(shards) == 0 {
		return nil, fmt.Errorf("at least one shard is required")
	}

	copied := append([]Shard(nil), shards...)
	for i, shard := range copied {
		if shard.Name == "" {
			return nil, fmt.Errorf("shard %d has no name", i)
		}
	}
	return &Router{shards: copied}, nil
}

// ShardForCustomer uses a stable modulo rule shared by the seed loader and
// coordinator. IDs start at one, so customer 1 is placed on shard 0.
func (r *Router) ShardForCustomer(customerID int64) (Shard, error) {
	if customerID < 1 {
		return Shard{}, fmt.Errorf("customer ID must be greater than zero")
	}

	index := (customerID - 1) % int64(len(r.shards))
	return r.shards[index], nil
}

// ScatterPlan returns all nodes for queries that cannot be pruned.
func (r *Router) ScatterPlan() []Shard {
	return append([]Shard(nil), r.shards...)
}
