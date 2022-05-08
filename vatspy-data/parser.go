package vatspydata

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/vatsimnerd/util/mapupdate"
	"github.com/vatsimnerd/util/pubsub"
)

type parseState int

const (
	stReadCategory parseState = iota
	stReadCountries
	stReadAirports
	stReadFIRs
	stReadUIRs

	coordErrorTemplate = "invalid lat/lng value '%s' on line %d"
)

func (p *Provider) parseData(raw []byte) error {
	log.Info("parsing vatspy data")
	state := stReadCategory
	countries := make(map[string]Country)
	airports := make(map[string]AirportMeta)
	firs := make(map[string]FIR)
	uirs := make(map[string]UIR)

	sc := bufio.NewScanner(bytes.NewReader(raw))
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())

		if len(line) == 0 || line[0] == ';' {
			// skip comments and empty lines
			continue
		}

		tokens := strings.Split(line, "|")

	redecide:
		switch state {
		case stReadCategory:
			if line[0] == '[' {
				cat := line[1 : len(line)-1]
				cat = strings.ToLower(cat)
				switch cat {
				case "countries":
					state = stReadCountries
				case "airports":
					state = stReadAirports
				case "firs":
					state = stReadFIRs
				case "uirs":
					state = stReadUIRs
				}
			}
		case stReadCountries:
			if len(tokens) != 3 {
				state = stReadCategory
				goto redecide
			}
			country := Country{tokens[0], tokens[1], tokens[2]}
			countries[country.Prefix] = country
		case stReadAirports:
			if len(tokens) != 7 {
				state = stReadCategory
				goto redecide
			}

			lat, err := strconv.ParseFloat(tokens[2], 64)
			if err != nil {
				return fmt.Errorf(coordErrorTemplate, tokens[2], lineNum)
			}
			lng, err := strconv.ParseFloat(tokens[3], 64)
			if err != nil {
				return fmt.Errorf(coordErrorTemplate, tokens[3], lineNum)
			}

			airport := AirportMeta{
				ICAO:     tokens[0],
				Name:     tokens[1],
				Position: Point{Lat: lat, Lng: lng},
				IATA:     tokens[4],
				FIRID:    tokens[5],
				IsPseudo: tokens[6] == "1",
			}
			airports[airport.ICAO] = airport

		case stReadFIRs:
			if len(tokens) != 4 {
				state = stReadCategory
				goto redecide
			}

			fir := FIR{
				ID:       tokens[0],
				Name:     tokens[1],
				Prefix:   tokens[2],
				ParentID: tokens[3],
			}

			if bnds, found := p.bdrs[fir.ID]; found {
				fir.Boundaries = bnds
			} else if bnds, found := p.bdrs[fir.Prefix]; found {
				fir.Boundaries = bnds
			} else if bnds, found := p.bdrs[fir.ParentID]; found {
				fir.Boundaries = bnds
			}

			firs[fir.ID] = fir

		case stReadUIRs:
			if len(tokens) != 3 {
				state = stReadCategory
				goto redecide
			}

			firIDs := strings.Split(tokens[2], ",")

			uir := UIR{
				ID:     tokens[0],
				Name:   tokens[1],
				FIRIDs: firIDs,
			}
			uirs[uir.ID] = uir
		}
	}

	countriesSet, countriesDel := mapupdate.Update[Country, mapupdate.Comparable[Country]](p.countries, countries, &p.dataLock)
	for _, update := range pubsub.MakeUpdates(countriesSet, countriesDel, ObjectTypeCountry) {
		p.Notify(update)
	}

	airportsSet, airportsDel := mapupdate.Update[AirportMeta, mapupdate.Comparable[AirportMeta]](p.airports, airports, &p.dataLock)
	for _, update := range pubsub.MakeUpdates(airportsSet, airportsDel, ObjectTypeAirportMeta) {
		p.Notify(update)
	}

	firsSet, firsDel := mapupdate.Update[FIR, mapupdate.Comparable[FIR]](p.firs, firs, &p.dataLock)
	for _, update := range pubsub.MakeUpdates(firsSet, firsDel, ObjectTypeFIR) {
		p.Notify(update)
	}

	uirsSet, uirsDel := mapupdate.Update[UIR, mapupdate.Comparable[UIR]](p.uirs, uirs, &p.dataLock)
	for _, update := range pubsub.MakeUpdates(uirsSet, uirsDel, ObjectTypeUIR) {
		p.Notify(update)
	}
	p.Fin()

	return nil
}
