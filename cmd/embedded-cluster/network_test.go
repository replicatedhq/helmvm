package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateCIDR(t *testing.T) {
	for _, tt := range []struct {
		name string
		cidr string
		err  string
	}{
		{
			name: "valid cidr",
			cidr: "10.0.0.0/16",
		},
		{
			name: "small cidr",
			cidr: "10.0.0.0/24",
			err:  "cidr needs to be at least a /16",
		},
		{
			name: "invalid cidr",
			cidr: "500.0.0.0/16",
			err:  "invalid CIDR address",
		},
		{
			name: "a /32 cidr",
			cidr: "10.0.0.0/32",
			err:  "cidr needs to be at least a /16",
		},
		{
			name: "a public cidr",
			cidr: "100.0.0.0/16",
			err:  "cidr is not within the private ranges",
		},
		{
			name: "matching the whole private range",
			cidr: "192.168.0.0/16",
		},
		{
			name: "matching the whole 172 range",
			cidr: "172.16.0.0/12",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCIDR(tt.cidr); err != nil {
				assert.NotEmpty(t, tt.err, "received unexpected error")
				assert.Contains(t, err.Error(), tt.err, "unexpected error message")
				return
			}
			assert.Empty(t, tt.err, "unexpected error received")
		})
	}
}
