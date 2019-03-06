package main

import "testing"

type CritHeightAndCycleResult struct {
	p                  int
	c                  int
	expectedCritHeight int
	expectedCritCycle  int
}

var critHeightAndCycleResults = []CritHeightAndCycleResult{
	{5, 1, 0, 3},
	{5, 2, 2, 2},
	{5, 3, 2, 1},
	{5, 4, 0, 2},
}

//TestCritHeightAndCycle: critHeightAndCycle(p, c int) => (critHeight, critCycle int)
func TestCritHeightAndCycle(t *testing.T) {
	for _, test := range critHeightAndCycleResults {
		resultHeight, resultCycle := critHeightAndCycle(test.p, test.c)
		if resultHeight != test.expectedCritHeight || resultCycle != test.expectedCritCycle {
			t.Error("Check critHeightAndCycle function.")
		}
	}
}

type DynamicOperatorResult struct {
	p        int
	c        int
	value    int
	expected int
}

var dynamicOperatorResults = []DynamicOperatorResult{
	{7, 5, 4, 0},
}

//TestDynamicOperator: dynamicOperator(p, c, value int) => expected int
func TestDynamicOperator(t *testing.T) {
	for _, test := range dynamicOperatorResults {
		result := dynamicOperator(test.p, test.c, test.value)
		if result != test.expected {
			t.Error("Check dynamicOperator function.")
		}
	}
}
