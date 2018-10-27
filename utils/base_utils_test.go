package utils

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		ip       string
		port     uint16
		expected NetAddr
	}{
		{
			"1.1.1.1",
			45,
			NetAddr{ip: "1.1.1.1", port: 45},
		},
	}
	for i, test := range tests {
		na := New(test.ip, test.port)
		var nai interface{} = *na
		testNa, ok := nai.(NetAddr)
		if !ok {
			t.Errorf("Test %d: requested-netaddr should be type NetAddr", i)
		}
		if testNa.ip != test.ip {
			t.Errorf("Test %d: expected netaddr ip %s, result %s", i, test.ip, testNa.ip)
		}
		if testNa.port != test.port {
			t.Errorf("Test %d: expected netaddr port %v, result %v", i, test.ip, testNa.ip)
		}
	}
}
