package common

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

func TestJsonFloat64(t *testing.T) {
	tss := []struct {
		f        float64
		expected string
	}{
		{f: 1.2, expected: "1.2"},
		{f: 0.0000001234, expected: "0.0000001234"},
		{f: 1.000000, expected: "1"},
		{f: 0.000000, expected: "0"},
		{f: 0.001000, expected: "0.001"},
		{f: 10.0, expected: "10"},
	}
	for _, ts := range tss {
		t.Run(ts.expected, func(t *testing.T) {
			bs, err := json.Marshal(JsonFloat64(ts.f))
			if err != nil {
				t.Fatalf("Marshalling failed with %v", err)
			}
			if string(bs) != ts.expected {
				t.Fatalf("Expected marshalling of %f to be exactly '%v' but was '%v'", ts.f, ts.expected, string(bs))
			}
		})
	}
}

func TestJsonFloat64Fails(t *testing.T) {
	tss := []struct {
		f float64
	}{
		{f: math.Inf(1)},
		{f: math.NaN()},
	}
	for _, ts := range tss {
		t.Run(fmt.Sprintf("%f", ts.f), func(t *testing.T) {
			_, err := json.Marshal(JsonFloat64(ts.f))
			if err == nil {
				t.Fatal("Expected marshalling to fail")
			}
		})
	}
}
