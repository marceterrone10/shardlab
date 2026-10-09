package analyzer

import (
	"reflect"
	"testing"
)

func TestSuggestAggregatesAndGroupsFrequentJoins(t *testing.T) {
	observations := []JoinObservation{
		{LeftTable: "clientes", RightTable: "contactos", Count: 7},
		{LeftTable: "contactos", RightTable: "clientes", Count: 5},
		{LeftTable: "clientes", RightTable: "oportunidades", Count: 10},
		{LeftTable: "actividades", RightTable: "usuarios", Count: 3},
	}

	got, err := Suggest(observations, 10)
	if err != nil {
		t.Fatalf("Suggest() error = %v", err)
	}

	want := []Suggestion{{
		Tables: []string{"clientes", "contactos", "oportunidades"},
		Weight: 22,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Suggest() = %#v, want %#v", got, want)
	}
}

func TestSuggestRejectsInvalidThreshold(t *testing.T) {
	if _, err := Suggest(nil, 0); err == nil {
		t.Fatal("Suggest(nil, 0) returned no error")
	}
}
