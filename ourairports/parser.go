package ourairports

import (
	"fmt"
	"strconv"
	"strings"
)

func parseString(field string) (string, error) {
	if strings.HasPrefix(field, "\"") && strings.HasSuffix(field, "\"") && len(field) >= 2 {
		return field[1 : len(field)-1], nil
	}
	return "", fmt.Errorf("'%s' is not a doublequoted string", field)
}

func parseRunway(line string) (*Runway, *Runway, error) {
	var err error
	tokens := strings.Split(line, ",")

	icao, err := parseString(tokens[2])
	if err != nil {
		// ICAO is missing
		return nil, nil, fmt.Errorf("invalid runway line '%s'", line)
	}

	length, err := strconv.ParseInt(tokens[3], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway length '%s'", tokens[3])
	}

	width, err := strconv.ParseInt(tokens[4], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway width '%s'", tokens[4])
	}

	surface, _ := parseString(tokens[5])
	lighted := tokens[6] == "1"
	closed := tokens[7] == "1"

	leIdent, err := parseString(tokens[8])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway ident '%s'", tokens[8])
	}

	leLat, err := strconv.ParseFloat(tokens[9], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway latitude '%s'", tokens[9])
	}
	leLng, err := strconv.ParseFloat(tokens[10], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway longitude '%s'", tokens[10])
	}

	leElev, err := strconv.ParseInt(tokens[11], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway elevation '%s'", tokens[11])
	}
	leHdg, err := strconv.ParseFloat(tokens[12], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway heading %s", tokens[12])
	}

	heIdent, err := parseString(tokens[14])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway ident '%s'", tokens[14])
	}
	heLat, err := strconv.ParseFloat(tokens[15], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway latitude '%s'", tokens[15])
	}
	heLng, err := strconv.ParseFloat(tokens[16], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway longitude '%s'", tokens[16])
	}

	heElev, err := strconv.ParseInt(tokens[17], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway elevation '%s'", tokens[17])
	}
	heHdg, err := strconv.ParseFloat(tokens[18], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid runway heading '%s'", tokens[18])
	}

	rwy1 := &Runway{
		ICAO:        icao,
		LengthFt:    int(length),
		WidthFt:     int(width),
		Surface:     surface,
		Lighted:     lighted,
		Closed:      closed,
		Ident:       leIdent,
		Latitude:    leLat,
		Longitude:   leLng,
		ElevationFt: int(leElev),
		Heading:     leHdg,
		ActiveTO:    false,
		ActiveLnd:   false,
	}

	rwy2 := &Runway{
		ICAO:        icao,
		LengthFt:    int(length),
		WidthFt:     int(width),
		Surface:     surface,
		Lighted:     lighted,
		Closed:      closed,
		Ident:       heIdent,
		Latitude:    heLat,
		Longitude:   heLng,
		ElevationFt: int(heElev),
		Heading:     heHdg,
		ActiveTO:    false,
		ActiveLnd:   false,
	}

	return rwy1, rwy2, nil
}
