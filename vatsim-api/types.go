package vatsimapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type (
	Facility int

	Controller struct {
		Cid           int       `json:"cid"`
		Name          string    `json:"name"`
		Callsign      string    `json:"callsign"`
		Frequency     float64   `json:"frequency"`
		Facility      Facility  `json:"facility"`
		Rating        int       `json:"rating"`
		Server        string    `json:"server"`
		VisualRange   int       `json:"visual_range"`
		AtisCode      string    `json:"atis_code,omitempty"`
		TextAtis      string    `json:"text_atis"`
		LastUpdated   time.Time `json:"last_updated"`
		LogonTime     time.Time `json:"logon_time"`
		HumanReadable string    `json:"human_readable"`
	}

	Pilot struct {
		Cid         int         `json:"cid"`
		Name        string      `json:"name"`
		Callsign    string      `json:"callsign"`
		Server      string      `json:"server"`
		PilotRating int         `json:"pilot_rating"`
		Latitude    float64     `json:"latitude"`
		Longitude   float64     `json:"longitude"`
		Altitude    int         `json:"altitude"`
		Groundspeed int         `json:"groundspeed"`
		Transponder string      `json:"transponder"`
		Heading     int         `json:"heading"`
		QnhIHg      float64     `json:"qnh_i_hg"`
		QnhMb       int         `json:"qnh_mb"`
		FlightPlan  *FlightPlan `json:"flight_plan"`
		LogonTime   time.Time   `json:"logon_time"`
		LastUpdated time.Time   `json:"last_updated"`
	}
)

const (
	FacilityATIS     = 1
	FacilityDelivery = 2
	FacilityGround   = 3
	FacilityTower    = 4
	FacilityApproach = 5
	FacilityRadar    = 6

	dateLayout = "2006-01-02T15:04:05"
)

func (c Controller) NE(o Controller) bool {
	return c.Cid != o.Cid ||
		c.Name != o.Name ||
		c.Callsign != o.Callsign ||
		c.Frequency != o.Frequency ||
		c.Facility != o.Facility ||
		c.Rating != o.Rating ||
		c.Server != o.Server ||
		c.VisualRange != o.VisualRange ||
		c.AtisCode != o.AtisCode ||
		c.TextAtis != o.TextAtis ||
		c.LogonTime != o.LogonTime
}

func (p Pilot) NE(o Pilot) bool {
	if p.Cid != o.Cid ||
		p.Name != o.Name ||
		p.Callsign != o.Callsign ||
		p.PilotRating != o.PilotRating ||
		p.Latitude != o.Latitude ||
		p.Longitude != o.Longitude ||
		p.Altitude != o.Altitude ||
		p.Groundspeed != o.Groundspeed ||
		p.Transponder != o.Transponder ||
		p.Heading != o.Heading ||
		p.QnhIHg != o.QnhIHg ||
		p.QnhMb != o.QnhMb ||
		p.LogonTime != o.LogonTime {
		return true
	}

	if (p.FlightPlan == nil) != (o.FlightPlan != nil) {
		return true
	}

	if p.FlightPlan == nil {
		return false
	}

	return *(p.FlightPlan) != *(o.FlightPlan)
}

func parseFrequency(frequency string) (float64, error) {
	freq, err := strconv.ParseFloat(frequency, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid frequency: %v", err)
	}

	if freq < 110 || freq > 140 {
		return 0, fmt.Errorf("frequency out of bounds %v", freq)
	}
	return freq, nil
}

func makeController(v VController) (Controller, error) {
	freq, err := parseFrequency(v.Frequency)
	if err != nil {
		return Controller{}, err
	}

	logonTime, err := time.Parse(dateLayout, v.LogonTime[:19])
	if err != nil {
		return Controller{}, fmt.Errorf("error parsing logon_time %s: %v", v.LogonTime, err)
	}

	lastUpdated, err := time.Parse(dateLayout, v.LastUpdated[:19])
	if err != nil {
		return Controller{}, fmt.Errorf("error parsing last_updated %s: %v", v.LastUpdated, err)
	}

	tokens := strings.Split(v.Callsign, "_")
	postfix := tokens[len(tokens)-1]

	if postfix == "SUP" || postfix == "OBS" {
		return Controller{}, fmt.Errorf("SUP or OBS callsign")
	}

	textAtis := strings.Join(v.TextAtis, "\n")
	lowerAtisText := strings.ToLower(textAtis)
	if strings.Contains(lowerAtisText, "supervisor") {
		return Controller{}, fmt.Errorf("supervisor in atis text")
	}

	return Controller{
		Cid:         v.Cid,
		Name:        v.Name,
		Callsign:    v.Callsign,
		Frequency:   freq,
		Facility:    Facility(v.Facility),
		Rating:      v.Rating,
		Server:      v.Server,
		VisualRange: v.VisualRange,
		AtisCode:    v.AtisCode,
		TextAtis:    textAtis,
		LastUpdated: lastUpdated,
		LogonTime:   logonTime,
	}, nil
}

func makePilot(v VPilot) (Pilot, error) {
	logonTime, err := time.Parse(dateLayout, v.LogonTime[:19])
	if err != nil {
		return Pilot{}, fmt.Errorf("error parsing logon_time %s: %v", v.LogonTime, err)
	}

	lastUpdated, err := time.Parse(dateLayout, v.LastUpdated[:19])
	if err != nil {
		return Pilot{}, fmt.Errorf("error parsing last_updated %s: %v", v.LastUpdated, err)
	}

	return Pilot{
		Cid:         v.Cid,
		Name:        v.Name,
		Callsign:    v.Callsign,
		Server:      v.Server,
		PilotRating: v.PilotRating,
		Latitude:    v.Latitude,
		Longitude:   v.Longitude,
		Altitude:    v.Altitude,
		Groundspeed: v.Groundspeed,
		Transponder: v.Transponder,
		Heading:     v.Heading,
		QnhIHg:      v.QnhIHg,
		QnhMb:       v.QnhMb,
		FlightPlan:  v.FlightPlan,
		LogonTime:   logonTime,
		LastUpdated: lastUpdated,
	}, nil
}
