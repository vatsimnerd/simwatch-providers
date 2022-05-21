package merged

import (
	"regexp"
	"testing"

	"github.com/vatsimnerd/util/set"
)

type testcase struct {
	name             string
	atis             string
	landingRunways   *set.Set[string]
	departureRunways *set.Set[string]
}

var (
	testcases = []testcase{
		{
			name:             "LFPG",
			atis:             "BONJOUR. THIS IS CHARLES DE GAULLE INFORMATION YANKEE RECORDED AT 1 6 4 3 U T C. LANDING RUNWAY 26 LEFT AND 27 RIGHT, TAKEOFF RUNWAY 26 RIGHT AND 27 LEFT. EXPECTED APPROACH ILS. EXPECTED DEPARTURES 5 ALPHA , 5 BRAVO , 5 ZULU. TRANSITION LEVEL 6 0. AFTER VACATING THE OUTER RUNWAY, HOLD SHORT OF THE INNER RUNWAY. BIRD ACTIVITY REPORTED. WIND 2 6 0 DEGREES, 9 KNOTS. VISIBILITY 1 0 KILOMETERS. CLOUDS SCATTERED 1800 FEET. TEMPERATURE 8, DEW POINT 5. Q N H 1 0 0 1, Q F E 0 9 8 7. CONFIRM ON FIRST CONTACT THAT YOU HAVE RECEIVED INFORMATION YANKEE.",
			landingRunways:   set.FromList([]string{"26L", "27R"}),
			departureRunways: set.FromList([]string{"27L", "26R"}),
		},
		{
			name:             "EDDV",
			atis:             "HANNOVER INFORMATION A MET REPORT TIME 1720 EXPECT ILS Z APPROACH RUNWAY 27C 27L OR 27R RUNWAYS IN USE 27C 27L AND 27R TRL 70 WIND 270 DEGREES 22 KNOTS GUSTS UP TO 33 KNOTS VISIBILITY 10 KILOMETERS IN THE VICINITY SHOWER CLOUDS BROKEN 2400 FEET TEMPERATURE 7 DEW POINT 3 QNH 985 TREND NOSIG HANNOVER INFORMATION A OUT",
			landingRunways:   set.FromList([]string{"27C", "27L", "27R"}),
			departureRunways: set.FromList([]string{"27C", "27L", "27R"}),
		},
		{
			name:             "EDDH",
			atis:             "HAMBURG INFORMATION E MET REPORT TIME 1720 EXPECT ILS APPROACH RUNWAY 23 RUNWAY 23 IN USE FOR LANDING AND TAKE OFF TRL 70 WHEN PASSING 2000 FEET CONTACT BREMEN RADAR ON FREQUENCY 123.925 WIND 240 DEGREES 25 KNOTS GUSTS UP TO 37 KNOTS VARIABLE BETWEEN 210 AND 270 DEGREES VISIBILITY 10 KILOMETERS LIGHT SHOWERS OF RAIN CLOUDS BROKEN CB 1800 FEET TEMPERATURE 6 DEW POINT 3 QNH 978 TREND TEMPORARY WIND 250 DEGREES 25 KNOTS GUSTS UP TO 45 KNOTS MODERATE SHOWERS OF RAIN INFORMATION E OUT",
			landingRunways:   set.FromList([]string{"23"}),
			departureRunways: set.FromList([]string{"23"}),
		},
		{
			name:             "EKCH",
			atis:             "THIS IS KASTRUP AIRPORT DEPARTURE AND ARRIVAL INFO W METREPORT 1720 EXPECT ILS APPROACH VISUAL APPROACH ON REQUEST ARRIVAL RUNWAY 22L AFTER LANDING VACATE RUNWAY DEPARTURE RUNWAY 22R TRANSITION LEVEL 75 WIND 200 DEGREES 19 KNOTS VISIBILITY MORE THAN 10 KILOMETERS LIGHT RAIN SKY CONDITION OVERCAST 1400 FEET TEMPERATURE 7 DEW POINT 5 QNH 974 TEMPORARY SKY CONDITION BROKEN 800 FEET IF UNABLE TO FOLLOW SID ADVICE ON INITIAL CONTACT SQUAWKMODE CHARLIE ON PUSHBACK THIS WAS KASTRUP AIRPORT DEPARTURE AND ARRIVAL INFO W",
			landingRunways:   set.FromList([]string{"22L"}),
			departureRunways: set.FromList([]string{"22R"}),
		},
		{
			name: "EGKK",
			atis: `GATWICK INFORMATION C TIME 142 0, RUNWAY IN USE 26L
			TRANSITION LEVEL FLIGHT LEVEL 7 0, SURFACE WIND 23 0, 6 KNOTS
			VISIBILITY 10KM OR MORE FEW 1 THOUSAND 7 HUNDRED FEET
			SCATTERED 2 THOUSAND 8 HUNDRED FEET TEMPERATURE +1 5,
			DEW POINT +1 1, QNH 101 8, ACKNOWLEDGE RECEIPT OF INFORMATION C
			AND ADVISE AIRCRAFT TYPE ON FIRST CONTACT`,
			landingRunways:   set.FromList([]string{"26L"}),
			departureRunways: set.FromList([]string{"26L"}),
		},
		{
			name: "EDDK",
			atis: `KOELN BONN INFORMATION C MET REPORT TIME 1450
			AUTOMATED WEATHER MESSAGE EXPECT ILS APPROACH RUNWAY 24
			RUNWAYS IN USE 14R AND 24 TRL 60 WIND 220 DEGREES 8
			KNOTS VISIBILITY 10 KILOMETERS MODERATE
			SHOWERS OF RAIN NO CLOUD BASE AVAILABLE TEMPERATURE 23
			DEW POINT 17 QNH 1015 TREND TEMPORARY WIND 240
			DEGREES 15 KNOTS GUSTS UP TO 25 KNOTS MODERATE
			SHOWERS OF RAIN CLOUDS BROKEN CB 3000 FEET INFORMATION C
			OUT`,
			landingRunways:   set.FromList([]string{"24"}),
			departureRunways: set.FromList([]string{"14R", "24"}),
		},
		{
			name: "EGCN",
			atis: `DONCASTER INFORMATION G TIME 105 0, RUNWAY IN USE 2 0,
			ILS APPROACH TO BE EXPECTED SURFACE WIND 25 0, 9 KNOTS VARYING
			BETWEEN 23 0, AND 290 DEGREES VISIBILITY 10KM OR MORE
			SCATTERED 3 THOUSAND 3 HUNDRED FEET TEMPERATURE +1 6,
			DEW POINT + 8, QNH 102 0, RUNWAY 2 0, DRY DRY DRY
			TURBULENCE MAY BE ENCOUNTERED IN THE FINAL STAGES OF THE APPROACH
			DEPARTING AIRCRAFT MAKE INITIAL CONTACT ON FREQUENCY 128.77 5,
			INCREASED BIRD ACTIVITY WITHIN THE AERODROME BOUNDARY
			ACKNOWLEDGE RECEIPT OF INFORMATION G AND ADVISE AIRCRAFT TYPE
			ON FIRST CONTACT`,
			landingRunways:   set.FromList([]string{"20"}),
			departureRunways: set.FromList([]string{"20"}),
		},
	}
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

func TestDetectRunways(t *testing.T) {
	for _, tc := range testcases {
		landing := detectArrivalRunways(tc.atis)
		if !landing.Eq(tc.landingRunways) {
			t.Errorf("[%s] landing runways don't match, expected %s, got %s",
				tc.name, tc.landingRunways, landing)
		}
		departure := detectDepartureRunways(tc.atis)
		if !departure.Eq(tc.departureRunways) {
			t.Errorf("[%s] departure runways don't match, expected %s, got %s",
				tc.name, tc.departureRunways, departure)
		}
	}
}
