package goshorewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

const rules01 = `
#ACTION		SOURCE		DEST		PROTO		DPORT	SPORT   ORIGDEST
ACCEPT		out		fw		tcp		1234
ACCEPT		out		fw		udp		51820
Ping(ACCEPT)	out		fw

DROP            banana:192.168.99.100   all:192.168.0.0

DNAT		out	    pve:10.90.0.4:22     tcp	    2222    -   
DNAT		out	    l:192.168.0.23	  tcp	    80      -    
DNAT		out	    l:192.168.1.5	  udp       51831   -     # K8s
REDIRECT	out	    51833	          udp       51833   -    

DNAT		h,h4,g	    pve:10.90.0.4:22     tcp	    2222    -   &ppp0
DNAT		h,h4,g	    l:192.168.0.23	  tcp	    80      -   &ppp0 
DNAT		h,h4,g	    l:192.168.0.23	  tcp,udp   443     -   &ppp0 
DNAT		h,h4,g	    l:192.168.5.5	  udp       51831   -   &ppp0  # asldk
REDIRECT	h,h4,g	    51833	          udp       51833   -   &ppp0 
`

func TestParseRules(t *testing.T) {
	rules := parseRules([]byte(rules01))
	assert.Equal(t, 13, len(rules), "expected 26 rules")

	assert.Equal(t, "ACCEPT", rules[0].Action)
	assert.Equal(t, "out", rules[0].Source)
	assert.Equal(t, "fw", rules[0].Destination)
	assert.Equal(t, "tcp", rules[0].Protocol)
	assert.Equal(t, "1234", rules[0].Dport)

	assert.Equal(t, "ACCEPT", rules[1].Action)
	assert.Equal(t, "out", rules[1].Source)
	assert.Equal(t, "fw", rules[1].Destination)
	assert.Equal(t, "udp", rules[1].Protocol)
	assert.Equal(t, "51820", rules[1].Dport)

	assert.Equal(t, "Ping(ACCEPT)", rules[2].Action)
	assert.Equal(t, "out", rules[2].Source)
	assert.Equal(t, "fw", rules[2].Destination)

	assert.Equal(t, "DROP", rules[3].Action)
	assert.Equal(t, "banana:192.168.99.100", rules[3].Source)
	assert.Equal(t, "all:192.168.0.0", rules[3].Destination)

	assert.Equal(t, "DNAT", rules[4].Action)
	assert.Equal(t, "out", rules[4].Source)
	assert.Equal(t, "DNAT", rules[4].Action)
	assert.Equal(t, "out", rules[4].Source)
	assert.Equal(t, "pve:10.90.0.4:22", rules[4].Destination)
	assert.Equal(t, "tcp", rules[4].Protocol)
	assert.Equal(t, "2222", rules[4].Dport)

	assert.Equal(t, "DNAT", rules[5].Action)
	assert.Equal(t, "out", rules[5].Source)
	assert.Equal(t, "l:192.168.0.23", rules[5].Destination)
	assert.Equal(t, "tcp", rules[5].Protocol)
	assert.Equal(t, "80", rules[5].Dport)

	assert.Equal(t, "DNAT", rules[6].Action)
	assert.Equal(t, "out", rules[6].Source)
	assert.Equal(t, "l:192.168.1.5", rules[6].Destination)
	assert.Equal(t, "udp", rules[6].Protocol)
	assert.Equal(t, "51831", rules[6].Dport)

	assert.Equal(t, "REDIRECT", rules[7].Action)
	assert.Equal(t, "out", rules[7].Source)
	assert.Equal(t, "51833", rules[7].Destination)
	assert.Equal(t, "udp", rules[7].Protocol)
	assert.Equal(t, "51833", rules[7].Dport)

	assert.Equal(t, "DNAT", rules[8].Action)
	assert.Equal(t, "h,h4,g", rules[8].Source)
	assert.Equal(t, "pve:10.90.0.4:22", rules[8].Destination)
	assert.Equal(t, "tcp", rules[8].Protocol)
	assert.Equal(t, "2222", rules[8].Dport)
	assert.Equal(t, "-", rules[8].Sport)
	assert.Equal(t, "&ppp0", rules[8].Origdest)

	assert.Equal(t, "DNAT", rules[9].Action)
	assert.Equal(t, "h,h4,g", rules[9].Source)
	assert.Equal(t, "l:192.168.0.23", rules[9].Destination)
	assert.Equal(t, "tcp", rules[9].Protocol)
	assert.Equal(t, "80", rules[9].Dport)
	assert.Equal(t, "-", rules[9].Sport)
	assert.Equal(t, "&ppp0", rules[9].Origdest)

	assert.Equal(t, "DNAT", rules[10].Action)
	assert.Equal(t, "h,h4,g", rules[10].Source)
	assert.Equal(t, "l:192.168.0.23", rules[10].Destination)
	assert.Equal(t, "tcp,udp", rules[10].Protocol)
	assert.Equal(t, "443", rules[10].Dport)
	assert.Equal(t, "-", rules[10].Sport)
	assert.Equal(t, "&ppp0", rules[10].Origdest)

	assert.Equal(t, "DNAT", rules[11].Action)
	assert.Equal(t, "h,h4,g", rules[11].Source)
	assert.Equal(t, "l:192.168.5.5", rules[11].Destination)
	assert.Equal(t, "udp", rules[11].Protocol)
	assert.Equal(t, "51831", rules[11].Dport)
	assert.Equal(t, "-", rules[11].Sport)
	assert.Equal(t, "&ppp0", rules[11].Origdest)

	assert.Equal(t, "REDIRECT", rules[12].Action)
	assert.Equal(t, "h,h4,g", rules[12].Source)
	assert.Equal(t, "51833", rules[12].Destination)
	assert.Equal(t, "udp", rules[12].Protocol)
	assert.Equal(t, "51833", rules[12].Dport)
	assert.Equal(t, "-", rules[12].Sport)
	assert.Equal(t, "&ppp0", rules[12].Origdest)

}
