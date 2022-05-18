package merged

import (
	"regexp"
	"testing"
)

func TestRunwayIdentExpr(t *testing.T) {
	re := regexp.MustCompile(runwayIdentExpr)
	match := re.FindAllStringSubmatch("35 LEFT", -1)
	if match[0][1] != "35 LEFT" {
		t.Errorf("Expected '35 LEFT', got %s", match[0])
	}

	match = re.FindAllStringSubmatch("22 RIGHT", -1)
	if match[0][1] != "22 RIGHT" {
		t.Errorf("Expected '22 RIGHT', got %s", match[0])
	}

	match = re.FindAllStringSubmatch("05 09 CENTER", -1)
	if match[0][1] != "05" {
		t.Errorf("Expected '05', got %s", match[0])
	}
	if match[0][2] != "09 CENTER" {
		t.Errorf("Expected '09 CENTER', got %s", match[0])
	}
}

func TestNormalizeRunwayIdent(t *testing.T) {
	var ident string

	type testcase struct {
		src string
		exp string
	}

	var testcases = []testcase{
		{"35L", "35L"},
		{"22", "22"},
		{"01 CENTER", "01C"},
	}

	for _, tc := range testcases {
		ident = normalizeIdent(tc.src)
		if ident != tc.exp {
			t.Errorf("Unexpected normalized value: expected %s, got %s", tc.exp, ident)
		}
	}
}
