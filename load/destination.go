package load

import (
	"fmt"

	"github.com/marceterrone10/shardlab/coordinator"
	"github.com/marceterrone10/shardlab/seed"
)

// Destination returns the shard that stores a customer and every child row in
// the bundle. Child rows follow bundle.Customer.ID.
func Destination(bundle seed.CustomerBundle, router *coordinator.Router) (string, error) {
	if router == nil {
		return "", fmt.Errorf("router is required")
	}
	if err := bundleBelongsToCustomer(bundle); err != nil {
		return "", err
	}

	shard, err := router.ShardForCustomer(bundle.Customer.ID)
	if err != nil {
		return "", err
	}
	return shard.Name, nil
}

func bundleBelongsToCustomer(bundle seed.CustomerBundle) error {
	customerID := bundle.Customer.ID
	for _, contact := range bundle.Contacts {
		if contact.CustomerID != customerID {
			return fmt.Errorf("contact %d belongs to customer %d, not %d", contact.ID, contact.CustomerID, customerID)
		}
	}
	for _, opportunity := range bundle.Opportunities {
		if opportunity.CustomerID != customerID {
			return fmt.Errorf("opportunity %d belongs to customer %d, not %d", opportunity.ID, opportunity.CustomerID, customerID)
		}
	}
	for _, activity := range bundle.Activities {
		if activity.CustomerID != customerID {
			return fmt.Errorf("activity %d belongs to customer %d, not %d", activity.ID, activity.CustomerID, customerID)
		}
	}
	return nil
}
