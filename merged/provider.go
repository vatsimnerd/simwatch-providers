package merged

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	vatsimapi "github.com/vatsimnerd/simwatch-providers/vatsim-api"
	vatspydata "github.com/vatsimnerd/simwatch-providers/vatspy-data"
	"github.com/vatsimnerd/util/pubsub"
)

var (
	log = logrus.WithField("module", "merged")
)

type Provider struct {
	*pubsub.Provider

	apiConfig  *vatsimapi.Config
	dataConfig *vatspydata.Config

	stop    chan bool
	stopped bool

	airports     map[string]Airport
	radars       map[string]Radar
	pilots       map[string]Pilot
	airportsIata map[string]Airport

	countries  map[string]vatspydata.Country
	firs       map[string]vatspydata.FIR
	firsPrefix map[string]vatspydata.FIR
	uirs       map[string]vatspydata.UIR

	dataLock sync.RWMutex
}

const (
	ObjectTypeAirport pubsub.ObjectType = 200 + iota
	ObjectTypeRadar
	ObjectTypePilot
)

var (
	errNotFound = fmt.Errorf("not found")
)

func New(apiConfig *vatsimapi.Config, dataConfig *vatspydata.Config) *Provider {
	return &Provider{
		Provider: pubsub.NewProvider(),
		stop:     make(chan bool),
		stopped:  false,

		apiConfig:  apiConfig,
		dataConfig: dataConfig,

		airports:     make(map[string]Airport),
		radars:       make(map[string]Radar),
		pilots:       make(map[string]Pilot),
		airportsIata: make(map[string]Airport),

		countries:  make(map[string]vatspydata.Country),
		firs:       make(map[string]vatspydata.FIR),
		firsPrefix: make(map[string]vatspydata.FIR),
		uirs:       make(map[string]vatspydata.UIR),
	}
}

func (p *Provider) Start() error {
	if p.stopped {
		log.Fatal("can't start once stopped provider")
	}
	go p.loop()
	return nil
}

func (p *Provider) Stop() {
	p.stop <- true
}

func (p *Provider) loop() {
	static := vatspydata.New(p.dataConfig)
	ssub := static.Subscribe(1024)

	dynamic := vatsimapi.New(p.apiConfig)
	dsub := dynamic.Subscribe(1024)
	dynamicStarted := false

	static.Start()
	defer static.Stop()

	p.SetInitialNotifier(func(sub pubsub.Subscription) {
		// initial notifier may take time and ponentially
		// fill up the notification chan so we're gonna make it
		// async to allow chan reading in another thread
		go func() {
			log.Debug("running initial notifier")
			p.dataLock.RLock()
			defer p.dataLock.RUnlock()
			for _, arpt := range p.airports {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeAirport, Obj: arpt})
			}
			for _, pilot := range p.pilots {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypePilot, Obj: pilot})
			}
			for _, radar := range p.radars {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeRadar, Obj: radar})
			}
		}()
	})

loop:
	for {
		select {
		case upd := <-ssub.Updates():
			log.WithFields(logrus.Fields{
				"uType": upd.UType,
				"oType": upd.OType,
				"obj":   upd.Obj,
			}).Debug("got update from vatspy data provider")
			switch upd.UType {
			case pubsub.UpdateTypeFin:
				// static data is ready, starting dynamic
				log.Info("initial static data ready, starting dynamic provider")
				if !dynamicStarted {
					p.SetDataReady(true)
					dynamic.Start()
					defer dynamic.Stop()
					dynamicStarted = true
				}

			case pubsub.UpdateTypeSet:
				switch upd.OType {
				case vatspydata.ObjectTypeCountry:
					country, ok := upd.Obj.(vatspydata.Country)
					if !ok {
						log.Error("object is expected to be Country, got %T", upd.Obj)
						continue
					}
					p.setCountry(country)
				case vatspydata.ObjectTypeAirportMeta:
					am, ok := upd.Obj.(vatspydata.AirportMeta)
					if !ok {
						log.Error("object is expected to be AirportMeta, got %T", upd.Obj)
						continue
					}
					p.setAirport(am)
				case vatspydata.ObjectTypeFIR:
					fir, ok := upd.Obj.(vatspydata.FIR)
					if !ok {
						log.Error("object is expected to be FIR, got %T", upd.Obj)
						continue
					}
					p.setFIR(fir)
				case vatspydata.ObjectTypeUIR:
					uir, ok := upd.Obj.(vatspydata.UIR)
					if !ok {
						log.Error("object is expected to be UIR, got %T", upd.Obj)
						continue
					}
					p.setUIR(uir)
				}

			case pubsub.UpdateTypeDelete:
				switch upd.OType {
				case vatspydata.ObjectTypeCountry:
					country, ok := upd.Obj.(vatspydata.Country)
					if !ok {
						log.Error("object is expected to be Country, got %T", upd.Obj)
						continue
					}
					p.deleteCountry(country)
				case vatspydata.ObjectTypeAirportMeta:
					am, ok := upd.Obj.(vatspydata.AirportMeta)
					if !ok {
						log.Error("object is expected to be AirportMeta, got %T", upd.Obj)
						continue
					}
					p.deleteAirport(am)
				case vatspydata.ObjectTypeFIR:
					fir, ok := upd.Obj.(vatspydata.FIR)
					if !ok {
						log.Error("object is expected to be FIR, got %T", upd.Obj)
						continue
					}
					p.deleteFIR(fir)
				case vatspydata.ObjectTypeUIR:
					uir, ok := upd.Obj.(vatspydata.UIR)
					if !ok {
						log.Error("object is expected to be UIR, got %T", upd.Obj)
						continue
					}
					p.deleteUIR(uir)
				}
			}
		case upd := <-dsub.Updates():
			log.WithFields(logrus.Fields{
				"uType": upd.UType,
				"oType": upd.OType,
				"obj":   upd.Obj,
			}).Debug("got update from vatsim api provider")
			switch upd.UType {
			case pubsub.UpdateTypeFin:
				p.Fin()
			case pubsub.UpdateTypeSet:
				switch upd.OType {
				case vatsimapi.ObjectTypePilot:
					pilot, ok := upd.Obj.(vatsimapi.Pilot)
					if !ok {
						log.Error("object is expected to be Pilot, got %T", upd.Obj)
						continue
					}
					p.setPilot(pilot)
				case vatsimapi.ObjectTypeController:
					ctrl, ok := upd.Obj.(vatsimapi.Controller)
					if !ok {
						log.Error("object is expected to be Controller, got %T", upd.Obj)
						continue
					}
					p.setController(ctrl)
				}
			case pubsub.UpdateTypeDelete:
				switch upd.OType {
				case vatsimapi.ObjectTypePilot:
					pilot, ok := upd.Obj.(vatsimapi.Pilot)
					if !ok {
						log.Error("object is expected to be Pilot, got %T", upd.Obj)
						continue
					}
					p.deletePilot(pilot)
				case vatsimapi.ObjectTypeController:
					ctrl, ok := upd.Obj.(vatsimapi.Controller)
					if !ok {
						log.Error("object is expected to be Controller, got %T", upd.Obj)
						continue
					}
					p.deleteController(ctrl)
				}
			}
		case <-p.stop:
			break loop
		}
	}
}

func (p *Provider) setCountry(c vatspydata.Country) {
	log.WithField("country", c.Prefix).Debug("setting country")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	p.countries[c.Prefix] = c
}

func (p *Provider) deleteCountry(c vatspydata.Country) {
	log.WithField("country", c.Prefix).Debug("deleting country")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	delete(p.countries, c.Prefix)
}

func (p *Provider) setFIR(f vatspydata.FIR) {
	log.WithField("fir", f.ID).Debug("setting fir")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	p.firs[f.ID] = f
	p.firsPrefix[f.Prefix] = f
}

func (p *Provider) deleteFIR(f vatspydata.FIR) {
	log.WithField("fir", f.ID).Debug("deleting fir")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	if fir, found := p.firs[f.ID]; found {
		delete(p.firs, fir.ID)
		delete(p.firsPrefix, fir.Prefix)
	}
}

func (p *Provider) setUIR(u vatspydata.UIR) {
	log.WithField("uir", u.ID).Debug("setting uir")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	p.uirs[u.ID] = u
}

func (p *Provider) deleteUIR(u vatspydata.UIR) {
	log.WithField("uir", u.ID).Debug("deleting uir")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	delete(p.uirs, u.ID)
}

func (p *Provider) setAirport(am vatspydata.AirportMeta) {
	log.WithField("arpt", am.ICAO).Debug("setting airport")
	var arpt Airport
	p.dataLock.Lock()
	defer p.dataLock.Unlock()

	if ex, found := p.airports[am.ICAO]; found {
		arpt = ex
		arpt.Meta = am
		delete(p.airports, ex.Meta.ICAO)
		delete(p.airportsIata, ex.Meta.IATA)
	} else {
		arpt = Airport{Meta: am}
	}

	p.airports[arpt.Meta.ICAO] = arpt
	p.airportsIata[arpt.Meta.IATA] = arpt
	update := pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeAirport, Obj: arpt}
	p.Notify(update)
}

func (p *Provider) deleteAirport(am vatspydata.AirportMeta) {
	log.WithField("arpt", am.ICAO).Debug("deleting airport")
	p.dataLock.Lock()
	defer p.dataLock.Unlock()

	if ex, found := p.airports[am.ICAO]; found {
		delete(p.airports, ex.Meta.ICAO)
		delete(p.airportsIata, ex.Meta.IATA)
		update := pubsub.Update{UType: pubsub.UpdateTypeDelete, OType: ObjectTypeAirport, Obj: ex}
		p.Notify(update)
	}
}

func (p *Provider) setController(c vatsimapi.Controller) {
	clog := log.WithField("callsign", c.Callsign)
	p.dataLock.Lock()
	defer p.dataLock.Unlock()

	tokens := strings.Split(c.Callsign, "_")
	prefix := tokens[0]

	if c.Facility == 0 {
		clog.Debug("skipping ctrl with facility=0")
		return
	} else if c.Facility >= 1 && c.Facility <= 5 {

		arpt, err := p.findAirportUnsafe(prefix)
		if err != nil {
			clog.Error("can't find airport for ctrl")
			return
		}

		alog := clog.WithField("icao", arpt.Meta.ICAO)

		switch c.Facility {
		case vatsimapi.FacilityATIS:
			arpt.Controllers.ATIS = &c
			c.HumanReadable = fmt.Sprintf("%s ATIS", arpt.Meta.Name)
			alog.Debug("atis set")
		case vatsimapi.FacilityDelivery:
			arpt.Controllers.Delivery = &c
			c.HumanReadable = fmt.Sprintf("%s Delivery", arpt.Meta.Name)
			alog.Debug("delivery set")
		case vatsimapi.FacilityGround:
			arpt.Controllers.Ground = &c
			c.HumanReadable = fmt.Sprintf("%s Ground", arpt.Meta.Name)
			alog.Debug("ground set")
		case vatsimapi.FacilityTower:
			arpt.Controllers.Tower = &c
			c.HumanReadable = fmt.Sprintf("%s Tower", arpt.Meta.Name)
			alog.Debug("tower set")
		case vatsimapi.FacilityApproach:
			arpt.Controllers.Approach = &c
			c.HumanReadable = fmt.Sprintf("%s Approach", arpt.Meta.Name)
			alog.Debug("approach set")
		}

		update := pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeAirport, Obj: arpt}
		p.Notify(update)
	} else if c.Facility == vatsimapi.FacilityRadar {
		clog.Debug("processing radar")

		firs := make(map[string]vatspydata.FIR, 0)
		var model *vatspydata.FIR

		clog.Debug("searching for firs")
		fir, err := p.findFIRUnsafe(prefix)
		if err == nil {
			firs[fir.ID] = fir
			model = &fir
		} else {
			uir, err := p.findUIRUnsafe(prefix)
			if err == nil {
				for _, firID := range uir.FIRIDs {
					fir, err := p.findFIRUnsafe(firID)
					if err == nil {
						firs[fir.ID] = fir
						if model == nil {
							model = &fir
						}
					} else {
						clog.WithFields(logrus.Fields{
							"fir": firID,
							"uir": uir.ID,
						}).Warn("can't find FIR provided by UIR")
					}
				}
			}
		}

		if len(firs) == 0 {
			clog.Error("can't find FIR or UIR for radar")
			return
		}

		controlName := "Centre"

		countryPrefix := model.ID[:2]
		if country, found := p.countries[countryPrefix]; found && country.ControlCustomName != "" {
			controlName = country.ControlCustomName
		}

		c.HumanReadable = fmt.Sprintf("%s %s", fir.Name, controlName)

		radar := Radar{Controller: c, FIRs: firs}
		p.radars[radar.Controller.Callsign] = radar

		update := pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeRadar, Obj: radar}
		p.Notify(update)

	} else {
		clog.WithField("facility", c.Facility).Error("invalid facility")
	}
}

func (p *Provider) deleteController(c vatsimapi.Controller) {
	p.dataLock.Lock()
	defer p.dataLock.Unlock()

	tokens := strings.Split(c.Callsign, "_")
	prefix := tokens[0]

	clog := log.WithField("callsign", c.Callsign)

	if c.Facility == 0 {
		clog.Debug("skipping ctrl with facility=0")
		return
	} else if c.Facility >= 1 && c.Facility <= 5 {

		arpt, err := p.findAirportUnsafe(prefix)
		if err != nil {
			clog.Error("can't find airport for ctrl")
			return
		}

		alog := clog.WithField("icao", arpt.Meta.ICAO)

		switch c.Facility {
		case vatsimapi.FacilityATIS:
			arpt.Controllers.ATIS = nil
			alog.Debug("atis removed")
		case vatsimapi.FacilityDelivery:
			arpt.Controllers.Delivery = nil
			alog.Debug("delivery removed")
		case vatsimapi.FacilityGround:
			arpt.Controllers.Ground = nil
			alog.Debug("ground removed")
		case vatsimapi.FacilityTower:
			arpt.Controllers.Tower = nil
			alog.Debug("tower removed")
		case vatsimapi.FacilityApproach:
			arpt.Controllers.Approach = nil
			alog.Debug("approach removed")
		}
		update := pubsub.Update{UType: pubsub.UpdateTypeDelete, OType: ObjectTypeAirport, Obj: arpt}
		p.Notify(update)
	} else if c.Facility == vatsimapi.FacilityRadar {
		if radar, found := p.radars[c.Callsign]; found {
			delete(p.radars, c.Callsign)
			update := pubsub.Update{UType: pubsub.UpdateTypeDelete, OType: ObjectTypeRadar, Obj: radar}
			p.Notify(update)
		}
	} else {
		clog.WithField("facility", c.Facility).Error("invalid facility")
	}
}

func (p *Provider) setPilot(vp vatsimapi.Pilot) {
	p.dataLock.Lock()
	defer p.dataLock.Unlock()

	pilot := Pilot{Pilot: vp}
	p.pilots[pilot.Callsign] = pilot
	update := pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypePilot, Obj: pilot}
	p.Notify(update)
}

func (p *Provider) deletePilot(vp vatsimapi.Pilot) {
	p.dataLock.Lock()
	defer p.dataLock.Unlock()
	if ex, found := p.pilots[vp.Callsign]; found {
		delete(p.pilots, vp.Callsign)
		update := pubsub.Update{UType: pubsub.UpdateTypeDelete, OType: ObjectTypePilot, Obj: ex}
		p.Notify(update)
	}
}

func (p *Provider) findAirportUnsafe(id string) (Airport, error) {
	if arpt, found := p.airports[id]; found {
		return arpt, nil
	}
	if arpt, found := p.airportsIata[id]; found {
		return arpt, nil
	}
	return Airport{}, errNotFound
}

func (p *Provider) findFIRUnsafe(id string) (vatspydata.FIR, error) {
	if fir, found := p.firs[id]; found {
		return fir, nil
	}
	if fir, found := p.firsPrefix[id]; found {
		return fir, nil
	}
	return vatspydata.FIR{}, errNotFound
}

func (p *Provider) findUIRUnsafe(id string) (vatspydata.UIR, error) {
	if uir, found := p.uirs[id]; found {
		return uir, nil
	}
	return vatspydata.UIR{}, errNotFound
}
