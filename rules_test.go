package goshorewall

import (
	"testing"
)

func TestRule_fillEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		input    Rule
		expected Rule
	}{
		{
			name: "Dport empty, Sport not",
			input: Rule{
				Sport: "8080",
			},
			expected: Rule{
				Dport: "-",
				Sport: "8080",
			},
		},
		{
			name: "Sport empty, Origdest not",
			input: Rule{
				Origdest: "1.2.3.4",
			},
			expected: Rule{
				Sport:    "-",
				Origdest: "1.2.3.4",
			},
		},
		{
			name: "Dport and Sport empty, Origdest not",
			input: Rule{
				Origdest: "1.2.3.4",
			},
			expected: Rule{
				Sport:    "-",
				Origdest: "1.2.3.4",
			},
		},
		{
			name: "Dport not empty, Sport empty, Origdest not",
			input: Rule{
				Dport:    "443",
				Origdest: "1.2.3.4",
			},
			expected: Rule{
				Dport:    "443",
				Sport:    "-",
				Origdest: "1.2.3.4",
			},
		},
		{
			name:     "All empty",
			input:    Rule{},
			expected: Rule{},
		},
		{
			name: "No fields empty",
			input: Rule{
				Action:      "ACCEPT",
				Source:      "net",
				Destination: "fw",
				Protocol:    "tcp",
				Dport:       "22",
				Sport:       "12345",
				Origdest:    "5.6.7.8",
			},
			expected: Rule{
				Action:      "ACCEPT",
				Source:      "net",
				Destination: "fw",
				Protocol:    "tcp",
				Dport:       "22",
				Sport:       "12345",
				Origdest:    "5.6.7.8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.fillEmpty()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestRule_Equals(t *testing.T) {
	testCases := []struct {
		name     string
		r1       Rule
		r2       Rule
		expected bool
	}{
		{
			name: "Identical rules",
			r1: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Dport: "22",
			},
			r2: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Dport: "22",
			},
			expected: true,
		},
		{
			name: "Different rules",
			r1: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw",
			},
			r2: Rule{
				Action: "DROP", Source: "net", Destination: "fw",
			},
			expected: false,
		},
		{
			name: "Equal after fillEmpty",
			r1: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Sport: "8080",
			},
			r2: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Dport: "-", Sport: "8080",
			},
			expected: true,
		},
		{
			name: "Not equal after fillEmpty",
			r1: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Sport: "8080",
			},
			r2: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw", Protocol: "tcp", Dport: "443", Sport: "8080",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.r1.Equals(tc.r2) != tc.expected {
				t.Errorf("Expected %v, got %v for r1.Equals(r2)", tc.expected, !tc.expected)
			}
		})
	}
}

func TestRule_Compare(t *testing.T) {
	testCases := []struct {
		name     string
		r1       Rule
		r2       Rule
		expected int // -1 for less, 0 for equal, 1 for greater
	}{
		{
			name: "Equal rules",
			r1: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw",
			},
			r2: Rule{
				Action: "ACCEPT", Source: "net", Destination: "fw",
			},
			expected: 0,
		},
		{
			name: "Action differs",
			r1: Rule{
				Action: "A",
			},
			r2: Rule{
				Action: "B",
			},
			expected: -1,
		},
		{
			name: "Source differs",
			r1: Rule{
				Action: "ACCEPT", Source: "a",
			},
			r2: Rule{
				Action: "ACCEPT", Source: "b",
			},
			expected: -1,
		},
		{
			name: "Equal after fillEmpty",
			r1: Rule{
				Action: "ACCEPT", Sport: "80",
			},
			r2: Rule{
				Action: "ACCEPT", Dport: "-", Sport: "80",
			},
			expected: 0,
		},
		{
			name: "Dport differs",
			r1: Rule{
				Action: "ACCEPT", Dport: "80",
			},
			r2: Rule{
				Action: "ACCEPT", Dport: "443",
			},
			expected: 1, // "80" > "443"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.r1.Compare(tc.r2)
			// Normalize result to -1, 0, or 1 to simplify comparison
			if result < 0 {
				result = -1
			} else if result > 0 {
				result = 1
			}

			if result != tc.expected {
				t.Errorf("Expected comparison result %d, got %d", tc.expected, result)
			}
		})
	}
}
