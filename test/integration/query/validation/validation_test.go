package validation

import (
	"fmt"
	"testing"

	"github.com/opencost/opencost-integration-tests/pkg/api"
)

func TestValidateResponseData(t *testing.T) {
	a := api.NewAPI()

	resp, err := a.GetCompute(api.ComputeRequest{
		Accumulate: "false",
		Window:     "1d",
		Aggregate:  "namespace",
		Idle:       "true",
		Step:       "1d",
	})

	if err != nil {
		t.Fatalf("Failed to get compute: %v", err)
	}

	if resp.Code != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.Code)
	}

	if len(resp.Data) == 0 || len(resp.Data[0]) == 0 {
		t.Fatal("Expected non-empty data response")
	}

	type TestCase struct {
		Name     string
		Check    func(v api.ComputeResponseItem) bool
		ErrorMsg string
	}

	testCases := []TestCase{
		{
			Name: "NegativeCost",
			Check: func(v api.ComputeResponseItem) bool {
				return v.CPUCost < 0 || v.GPUCost < 0 || v.RAMCost < 0 || v.TotalCost < 0 
			},
			ErrorMsg: "Negative cost detected",
		},
		{
			Name: "ZeroUsageNonZeroCost",
			Check: func(v api.ComputeResponseItem) bool {
				return (v.CPUCoreHours == 0 && v.CPUCost > 0) || (v.GPUHours == 0 && v.GPUCost > 0)
			},
			ErrorMsg: "Non-zero cost despite zero usage",
		},
		{
			Name: "TotalEfficiencyOutOfBounds",
			Check: func(v api.ComputeResponseItem) bool {
				return v.TotalEfficiency < 0 || v.TotalEfficiency > 1
			},
			ErrorMsg: "Total efficiency is out of bounds [0, 1]",
		},
		{
			Name: "StartAfterEnd",
			Check: func(v api.ComputeResponseItem) bool {
				return v.Start.After(v.End)
			},
			ErrorMsg: "Start time is after end time",
		},
		{
			Name: "NetworkUsageWithoutCost",
			Check: func(v api.ComputeResponseItem) bool {
				return v.NetworkTransferBytes > 0 && v.NetworkCost == 0
			},
			ErrorMsg: "Network usage reported but no cost",
		},
		{
			Name: "CostMismatch",
			Check: func(v api.ComputeResponseItem) bool {
				components := v.CPUCost + v.GPUCost + v.RAMCost + v.NetworkCost + v.LoadBalancerCost + v.SharedCost
				diff := components - v.TotalCost
				return diff > 0.01 || diff < -0.01 // Allowing float tolerance
			},
			ErrorMsg: "Total cost mismatch with component costs",
		},
	}

	for k, v := range resp.Data[0] {
		t.Run(fmt.Sprintf("Namespace=%s", k), func(t *testing.T) {
			for _, tc := range testCases {
				if tc.Check(v) {
					t.Errorf("[%s] %s: %+v", tc.Name, tc.ErrorMsg, v)
				}
			}
		})
	}
}
