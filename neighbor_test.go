package main

import "testing"

func TestNeighbors(t *testing.T) {
	p := point{10, 10}

	neighbors := p.neighbors()
	expected := []point{
		{9, 9},
		{10, 9},
		{11, 9},
		{9, 10},
		{11, 10},
		{9, 11},
		{10, 11},
		{11, 11},
	}

	for i := 0; i < len(expected); i++ {
		if neighbors[i] != expected[i] {
			t.Errorf("expected: % d got: % d", expected[i], neighbors[i])
		}
	}

}
