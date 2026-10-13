package coordinator

import "testing"

func TestRouterUsesStableModulo(t *testing.T) {
	router, err := NewRouter([]Shard{
		{Name: "shard0"},
		{Name: "shard1"},
		{Name: "shard2"},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	tests := []struct {
		customerID int64
		want       string
	}{
		{customerID: 1, want: "shard0"},
		{customerID: 2, want: "shard1"},
		{customerID: 3, want: "shard2"},
		{customerID: 4, want: "shard0"},
	}

	for _, test := range tests {
		shard, err := router.ShardForCustomer(test.customerID)
		if err != nil {
			t.Fatalf("ShardForCustomer(%d) error = %v", test.customerID, err)
		}
		if shard.Name != test.want {
			t.Errorf("ShardForCustomer(%d) = %q, want %q", test.customerID, shard.Name, test.want)
		}
	}
}

func TestRouterRejectsInvalidCustomerID(t *testing.T) {
	router, err := NewRouter([]Shard{{Name: "shard0"}})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	if _, err := router.ShardForCustomer(0); err == nil {
		t.Fatal("ShardForCustomer(0) returned no error")
	}
}
