package merged

import (
	vatsimapi "github.com/vatsimnerd/simwatch-providers/vatsim-api"
	vatspydata "github.com/vatsimnerd/simwatch-providers/vatspy-data"
)

type (
	Pilot struct {
		vatsimapi.Pilot
	}

	ControllerSet struct {
		ATIS     *vatsimapi.Controller `json:"atis"`
		Delivery *vatsimapi.Controller `json:"del"`
		Ground   *vatsimapi.Controller `json:"gnd"`
		Tower    *vatsimapi.Controller `json:"twr"`
		Approach *vatsimapi.Controller `json:"appr"`
	}

	Airport struct {
		Meta        vatspydata.AirportMeta `json:"meta"`
		Controllers ControllerSet          `json:"ctrls"`
	}

	Radar struct {
		Controller vatsimapi.Controller      `json:"ctrl"`
		FIRs       map[string]vatspydata.FIR `json:"firs"`
	}
)

func (p Pilot) NE(o Pilot) bool {
	return p.Pilot.NE(o.Pilot)
}

func (a Airport) NE(o Airport) bool {
	return a.Meta.NE(o.Meta) ||
		a.Controllers.NE(o.Controllers)
}

func (a Airport) IsControlled() bool {
	return !a.Controllers.IsEmpty()
}

func (cs ControllerSet) NE(o ControllerSet) bool {
	if (cs.ATIS == nil) != (o.ATIS == nil) || ((cs.ATIS != nil) && cs.ATIS.NE(*o.ATIS)) {
		return true
	}
	if (cs.Delivery == nil) != (o.Delivery == nil) || ((cs.Delivery != nil) && cs.Delivery.NE(*o.Delivery)) {
		return true
	}
	if (cs.Ground == nil) != (o.Ground == nil) || ((cs.Ground != nil) && cs.Ground.NE(*o.Ground)) {
		return true
	}
	if (cs.Tower == nil) != (o.Tower == nil) || ((cs.Tower != nil) && cs.Tower.NE(*o.Tower)) {
		return true
	}
	if (cs.Approach == nil) != (o.Approach == nil) || ((cs.Approach != nil) && cs.Approach.NE(*o.Approach)) {
		return true
	}
	return false
}

func (cs ControllerSet) IsEmpty() bool {
	return cs.ATIS == nil &&
		cs.Delivery == nil &&
		cs.Ground == nil &&
		cs.Tower == nil &&
		cs.Approach == nil
}

func (r Radar) NE(o Radar) bool {
	if r.Controller.NE(o.Controller) {
		return true
	}

	if len(r.FIRs) != len(o.FIRs) {
		return true
	}

	for id, fir := range r.FIRs {
		if ofir, found := o.FIRs[id]; !found || fir.NE(ofir) {
			return true
		}
	}

	return false
}
