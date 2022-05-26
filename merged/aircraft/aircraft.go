package aircraft

import (
	"encoding/json"
	"strconv"
)

type modelSource struct {
	ModelFullName       string
	Description         string
	WTC                 string
	WTG                 string
	Designator          string
	ManufacturerCode    string
	AircraftDescription string
	EngineCount         string
	EngineType          string
}

type AircraftType struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	WTC              string `json:"wtc"`
	WTG              string `json:"wtg"`
	Designator       string `json:"designator"`
	ManufacturerCode string `json:"manufacturer_code"`
	EngineCount      int    `json:"engine_count"`
	EngineType       string `json:"engine_type"`
}

var (
	AircraftTypes = make(map[string]AircraftType)
)

func init() {
	modelList := make([]modelSource, 0)
	err := json.Unmarshal([]byte(models), &modelList)
	if err != nil {
		panic(err)
	}

	for _, msrc := range modelList {
		ec, err := strconv.ParseInt(msrc.EngineCount, 10, 64)
		if err != nil {
			ec = 1
		}
		AircraftTypes[msrc.Designator] = AircraftType{
			Name:             msrc.ModelFullName,
			Description:      msrc.Description,
			WTC:              msrc.WTC,
			WTG:              msrc.WTG,
			Designator:       msrc.Designator,
			ManufacturerCode: msrc.ManufacturerCode,
			EngineCount:      int(ec),
			EngineType:       msrc.EngineType,
		}
	}
}
