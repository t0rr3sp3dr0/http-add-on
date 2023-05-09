package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	httpv1alpha1 "github.com/kedacore/http-add-on/operator/apis/http/v1alpha1"
	"github.com/kedacore/http-add-on/pkg/routing"
)

type testCase struct {
	name      string
	table     routing.Table
	counts    map[string]int
	retCounts map[string]int
}

func cases(r *require.Assertions) []testCase {
	return []testCase{
		{
			name: "empty queue",
			table: newRoutingTable(r, []hostAndTarget{
				{
					host:   "www.example.com",
					target: &httpv1alpha1.HTTPScaledObject{},
				},
				{
					host:   "www.example2.com",
					target: &httpv1alpha1.HTTPScaledObject{},
				},
			}),
			counts: make(map[string]int),
			retCounts: map[string]int{
				"www.example.com":  0,
				"www.example2.com": 0,
			},
		},
		{
			name: "one entry in queue, same entry in routing table",
			table: newRoutingTable(r, []hostAndTarget{
				{
					host:   "example.com",
					target: &httpv1alpha1.HTTPScaledObject{},
				},
			}),
			counts: map[string]int{
				"example.com": 1,
			},
			retCounts: map[string]int{
				"example.com": 1,
			},
		},
		{
			name: "one entry in queue, two in routing table",
			table: newRoutingTable(r, []hostAndTarget{
				{
					host:   "example.com",
					target: &httpv1alpha1.HTTPScaledObject{},
				},
				{
					host:   "example2.com",
					target: &httpv1alpha1.HTTPScaledObject{},
				},
			}),
			counts: map[string]int{
				"example.com": 1,
			},
			retCounts: map[string]int{
				"example.com":  1,
				"example2.com": 0,
			},
		},
	}
}
func TestGetHostCount(t *testing.T) {
	r := require.New(t)
	for _, tc := range cases(r) {
		for host, retCount := range tc.retCounts {
			t.Run(tc.name, func(t *testing.T) {
				r := require.New(t)
				ret, exists := getHostCount(
					host,
					tc.counts,
				)
				r.True(exists)
				r.Equal(retCount, ret)
			})
		}
	}
}

type hostAndTarget struct {
	host   string
	target *httpv1alpha1.HTTPScaledObject
}

func newRoutingTable(r *require.Assertions, entries []hostAndTarget) routing.Table {
	ret := newTestRoutingTable()
	for _, entry := range entries {
		ret.memory[entry.host] = entry.target
	}
	return ret
}

type testRoutingTable struct {
	memory map[string]*httpv1alpha1.HTTPScaledObject
}

func newTestRoutingTable() *testRoutingTable {
	return &testRoutingTable{
		memory: make(map[string]*httpv1alpha1.HTTPScaledObject),
	}
}

var _ routing.Table = (*testRoutingTable)(nil)

func (t testRoutingTable) Start(_ context.Context) error {
	return nil
}

func (t testRoutingTable) Route(req *http.Request) *httpv1alpha1.HTTPScaledObject {
	httpso, _ := t.memory[req.Host]
	return httpso
}

func (t testRoutingTable) HasSynced() bool {
	return true
}
