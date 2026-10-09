// Package seed creates reproducible CRM fixtures.
package seed

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Customers int64
	Seed      uint64
}

type Customer struct {
	ID        int64
	Name      string
	Email     string
	Status    string
	CreatedAt time.Time
}

type Contact struct {
	ID         int64
	CustomerID int64
	Name       string
	Email      string
	Role       string
	CreatedAt  time.Time
}

type Opportunity struct {
	ID         int64
	CustomerID int64
	Name       string
	Amount     string
	Status     string
	CloseDate  time.Time
	CreatedAt  time.Time
}

type Activity struct {
	ID          int64
	CustomerID  int64
	Type        string
	Description string
	OccurredAt  time.Time
}

type CustomerBundle struct {
	Customer      Customer
	Contacts      []Contact
	Opportunities []Opportunity
	Activities    []Activity
}

type Generator struct {
	options Options
}

func NewGenerator(options Options) (*Generator, error) {
	if options.Customers < 1 {
		return nil, fmt.Errorf("customer count must be greater than zero")
	}
	return &Generator{options: options}, nil
}

func (g *Generator) Customer(customerID int64) (CustomerBundle, error) {
	if customerID < 1 || customerID > g.options.Customers {
		return CustomerBundle{}, fmt.Errorf("customer ID %d is outside the configured range", customerID)
	}

	createdAt := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(value(g.options.Seed, customerID, 1)%365) * 24 * time.Hour)
	customer := Customer{
		ID:        customerID,
		Name:      fmt.Sprintf("Cliente %06d", customerID),
		Email:     fmt.Sprintf("cliente-%d@example.test", customerID),
		Status:    pick([]string{"activo", "prospecto", "inactivo"}, g.options.Seed, customerID, 2),
		CreatedAt: createdAt,
	}

	contactCount := 1 + int(value(g.options.Seed, customerID, 3)%3)
	contacts := make([]Contact, 0, contactCount)
	for i := 1; i <= contactCount; i++ {
		contacts = append(contacts, Contact{
			ID:         customerID*10 + int64(i),
			CustomerID: customerID,
			Name:       fmt.Sprintf("Contacto %d de cliente %d", i, customerID),
			Email:      fmt.Sprintf("contacto-%d-%d@example.test", customerID, i),
			Role:       pick([]string{"compras", "finanzas", "operaciones"}, g.options.Seed, customerID, uint64(10+i)),
			CreatedAt:  createdAt.Add(time.Duration(i) * time.Hour),
		})
	}

	opportunityCount := 1 + int(value(g.options.Seed, customerID, 20)%3)
	opportunities := make([]Opportunity, 0, opportunityCount)
	for i := 1; i <= opportunityCount; i++ {
		cents := 100_000 + value(g.options.Seed, customerID, uint64(20+i))%9_900_000
		opportunities = append(opportunities, Opportunity{
			ID:         customerID*10 + int64(i),
			CustomerID: customerID,
			Name:       fmt.Sprintf("Oportunidad %d de cliente %d", i, customerID),
			Amount:     fmt.Sprintf("%d.%02d", cents/100, cents%100),
			Status:     pick([]string{"abierta", "ganada", "perdida"}, g.options.Seed, customerID, uint64(30+i)),
			CloseDate:  createdAt.AddDate(0, 0, 30+i),
			CreatedAt:  createdAt.Add(time.Duration(i) * time.Hour),
		})
	}

	activityCount := 2 + int(value(g.options.Seed, customerID, 40)%4)
	activities := make([]Activity, 0, activityCount)
	for i := 1; i <= activityCount; i++ {
		activityType := pick([]string{"llamada", "email", "reunion"}, g.options.Seed, customerID, uint64(40+i))
		activities = append(activities, Activity{
			ID:          customerID*10 + int64(i),
			CustomerID:  customerID,
			Type:        activityType,
			Description: fmt.Sprintf("%s %d de cliente %d", activityType, i, customerID),
			OccurredAt:  createdAt.Add(time.Duration(i) * 24 * time.Hour),
		})
	}

	return CustomerBundle{
		Customer:      customer,
		Contacts:      contacts,
		Opportunities: opportunities,
		Activities:    activities,
	}, nil
}

func ParseSize(raw string) (int64, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return 0, fmt.Errorf("size cannot be empty")
	}

	multiplier := int64(1)
	switch normalized[len(normalized)-1] {
	case 'k':
		multiplier = 1_000
		normalized = normalized[:len(normalized)-1]
	case 'm':
		multiplier = 1_000_000
		normalized = normalized[:len(normalized)-1]
	}

	size, err := strconv.ParseInt(normalized, 10, 64)
	if err != nil || size < 1 {
		return 0, fmt.Errorf("invalid size %q", raw)
	}
	if size > (1<<63-1)/multiplier {
		return 0, fmt.Errorf("size %q is too large", raw)
	}
	return size * multiplier, nil
}

func pick(values []string, seed uint64, customerID int64, salt uint64) string {
	return values[value(seed, customerID, salt)%uint64(len(values))]
}

func value(seed uint64, customerID int64, salt uint64) uint64 {
	x := seed ^ uint64(customerID)*0x9e3779b97f4a7c15 ^ salt
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	return x ^ (x >> 31)
}
