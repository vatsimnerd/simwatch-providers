package vatspydata

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vatsimnerd/perfetch"
	"github.com/vatsimnerd/util/pubsub"
)

type Provider struct {
	*pubsub.Provider

	stop      chan bool
	stopped   bool
	bdrs      map[string]Boundaries
	countries map[string]Country
	firs      map[string]FIR
	uirs      map[string]UIR
	airports  map[string]AirportMeta

	dataLock sync.RWMutex
}

const (
	dataURL       = "https://raw.githubusercontent.com/vatsimnetwork/vatspy-data-project/master/VATSpy.dat"
	boundariesURL = "https://raw.githubusercontent.com/vatsimnetwork/vatspy-data-project/master/Boundaries.geojson"

	period  = 24 * time.Hour
	timeout = 5 * time.Second

	bootstrapRetries         = 3
	bootstrapRetriesCooldown = 3 * time.Second

	ObjectTypeCountry pubsub.ObjectType = 100 + iota
	ObjectTypeFIR
	ObjectTypeUIR
	ObjectTypeAirportMeta
)

func New() *Provider {
	return &Provider{
		Provider:  pubsub.NewProvider(),
		stop:      make(chan bool),
		stopped:   false,
		bdrs:      make(map[string]Boundaries),
		countries: make(map[string]Country),
		firs:      make(map[string]FIR),
		uirs:      make(map[string]UIR),
		airports:  make(map[string]AirportMeta),
	}
}

func (p *Provider) Start() {
	if p.stopped {
		log.Fatal("can't start once stopped provider")
	}
	go p.loop()
}

func (p *Provider) Stop() {
	p.stop <- true
}

func (p *Provider) loop() {
	log.Info("entering vatspy data provider loop()")

	p.SetInitialNotifier(func(sub pubsub.Subscription) {
		go func() {
			p.dataLock.Lock()
			defer p.dataLock.Unlock()

			for _, v := range p.countries {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeCountry, Obj: v})
			}
			for _, v := range p.airports {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeAirportMeta, Obj: v})
			}
			for _, v := range p.firs {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeFIR, Obj: v})
			}
			for _, v := range p.uirs {
				sub.Send(pubsub.Update{UType: pubsub.UpdateTypeSet, OType: ObjectTypeUIR, Obj: v})
			}
			sub.Fin()
		}()
	})

	bpoller := perfetch.New(period, perfetch.HTTPGetFetcher(boundariesURL, timeout))
	dpoller := perfetch.New(period, perfetch.HTTPGetFetcher(dataURL, timeout))

	bsub := bpoller.Subscribe(10)
	dsub := dpoller.Subscribe(10)

	r := 0
	for r < bootstrapRetries {
		err := bpoller.Start()
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
	defer bpoller.Stop()

	buf := <-bsub.Updates()
	err := p.parseBoundaries(buf)
	if err != nil {
		log.WithError(err).Fatal("error parsing boundaries, giving up")
	}

	r = 0
	for r < bootstrapRetries {
		err := dpoller.Start()
		if err == nil {
			break
		}
		r++
		log.WithError(err).WithField("retries_left", bootstrapRetries-r).Error("error fetching data (initial)")
		if r == bootstrapRetries {
			log.Fatal("error fetching data (initially), no retries left")
		}
		time.Sleep(bootstrapRetriesCooldown)
	}

	p.SetDataReady(true)
	defer dpoller.Stop()

loop:
	for {
		select {
		case buf := <-dsub.Updates():
			err := p.parseData(buf)
			if err != nil {
				log.WithError(err).Error("error parsing data")
			}
		case buf := <-bsub.Updates():
			err := p.parseBoundaries(buf)
			if err != nil {
				log.WithError(err).Error("error parsing boundaries")
			}
		case <-p.stop:
			p.stopped = true
			log.Info("stop signal received")
			break loop
		}
	}

	p.Dispose()
}
