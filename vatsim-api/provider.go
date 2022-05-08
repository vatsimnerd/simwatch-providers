package vatsimapi

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vatsimnerd/perfetch"
	"github.com/vatsimnerd/util/mapupdate"
	"github.com/vatsimnerd/util/pubsub"
)

type Provider struct {
	*pubsub.Provider

	stop    chan bool
	stopped bool

	controllers map[string]Controller
	pilots      map[string]Pilot

	dataLock sync.RWMutex
}

const (
	apiURL  = "https://data.vatsim.net/v3/vatsim-data.json"
	period  = 15 * time.Second
	timeout = 3 * time.Second

	bootstrapRetries         = 5
	bootstrapRetriesCooldown = 3 * time.Second

	ObjectTypeController pubsub.ObjectType = iota + 1
	ObjectTypePilot
)

func New() *Provider {
	return &Provider{
		Provider:    pubsub.NewProvider(),
		stop:        make(chan bool),
		stopped:     false,
		controllers: make(map[string]Controller),
		pilots:      make(map[string]Pilot),
	}
}

func (p *Provider) Start() error {
	if p.stopped {
		return fmt.Errorf("can't start once stopped provider")
	}
	go p.loop()
	return nil
}

func (p *Provider) Stop() {
	p.stop <- true
}

func (p *Provider) loop() {
	poller := perfetch.New(period, perfetch.HTTPGetFetcher(apiURL, timeout))
	psub := poller.Subscribe(1024)
	defer poller.Unsubscribe(psub)

	p.SetInitialNotifier(func(sub pubsub.Subscription) {
		// make notifier async to avoid reaching chan buffer limit
		go func() {
			p.dataLock.RLock()
			defer p.dataLock.RUnlock()
			for _, ctrl := range p.controllers {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeController, Obj: ctrl})
			}
			for _, pilot := range p.pilots {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypePilot, Obj: pilot})
			}
		}()
	})

	r := 0
	for r < bootstrapRetries {
		err := poller.Start()
		if err == nil {
			break
		}
		r++
		log.WithError(err).WithField("retries_left", bootstrapRetries-r).Error("error fetching boundaries (initial)")
		if r == bootstrapRetries {
			log.Fatal("error fetching boundaries (initially), no retries left")
		}
		time.Sleep(bootstrapRetriesCooldown)
	}
	defer poller.Stop()

loop:
	for {
		select {
		case raw := <-psub.Updates():
			log.Info("got update from vatsim api poller")
			data := Data{}

			err := json.Unmarshal(raw, &data)
			if err != nil {
				log.WithError(err).Error("error unmarshalling vatsim api data")
				continue loop
			}

			controllers := make(map[string]Controller)
			for _, vctrl := range data.Controllers {
				ctrl, err := makeController(vctrl)
				if err != nil {
					log.WithError(err).WithField("callsign", vctrl.Callsign).Debug("skipping invalid controller")
					continue
				}
				controllers[ctrl.Callsign] = ctrl
			}

			for _, vctrl := range data.ATIS {
				vctrl.Facility = FacilityATIS
				ctrl, err := makeController(vctrl)
				if err != nil {
					log.WithError(err).WithField("callsign", vctrl.Callsign).Debug("skipping invalid controller")
					continue
				}
				controllers[ctrl.Callsign] = ctrl
			}

			pilots := make(map[string]Pilot)
			for _, vpilot := range data.Pilots {
				pilot, err := makePilot(vpilot)
				if err != nil {
					log.WithError(err).WithField("callsign", vpilot.Callsign).Debug("skipping invalid pilot")
					continue
				}
				pilots[pilot.Callsign] = pilot
			}

			ctrlSet, ctrlDel := mapupdate.Update[Controller, mapupdate.Comparable[Controller]](p.controllers, controllers, &p.dataLock)
			for _, update := range pubsub.MakeUpdates(ctrlSet, ctrlDel, ObjectTypeController) {
				p.Notify(update)
			}

			pilotSet, pilotDel := mapupdate.Update[Pilot, mapupdate.Comparable[Pilot]](p.pilots, pilots, &p.dataLock)
			for _, update := range pubsub.MakeUpdates(pilotSet, pilotDel, ObjectTypePilot) {
				p.Notify(update)
			}

			p.Fin()

			p.SetDataReady(true)

		case <-p.stop:
			break loop
		}
	}

	p.Dispose()
}
