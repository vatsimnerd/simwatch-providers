package ourairports

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/perfetch"
	"github.com/vatsimnerd/util/pubsub"
)

type Provider struct {
	*pubsub.Provider

	cfg *Config

	stop    chan bool
	stopped bool

	runways map[string]map[string]*Runway

	dataLock sync.RWMutex
}

var (
	log = logrus.WithField("module", "ourairports")
)

const (
	OurairportsRunwaysURL = "https://ourairports.com/data/runways.csv"

	ObjecTypeRunway pubsub.ObjectType = 300 + iota
)

func New(cfg *Config) *Provider {
	return &Provider{
		Provider: pubsub.NewProvider(),
		cfg:      cfg,
		stop:     make(chan bool),
		stopped:  false,
		runways:  make(map[string]map[string]*Runway),
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
	defer p.Dispose()

	var rawChan <-chan []byte

	if strings.HasPrefix(p.cfg.URL, "http") {
		poller := perfetch.New(
			p.cfg.Poll.Period,
			perfetch.HTTPGetFetcher(p.cfg.URL, p.cfg.Poll.Timeout),
		)
		psub := poller.Subscribe(1024)
		defer poller.Unsubscribe(psub)

		p.SetInitialNotifier(func(sub pubsub.Subscription) {
			// make notifier async to avoid reaching chan buffer limit
			go func() {
				p.dataLock.RLock()
				defer p.dataLock.RUnlock()
				for _, rwmap := range p.runways {
					for _, rwy := range rwmap {
						update := pubsub.Update{
							UType: pubsub.UpdateTypeSet,
							OType: ObjecTypeRunway,
							Obj:   *rwy,
						}
						p.Notify(update)
					}
				}
			}()
		})

		r := 0
		for r < p.cfg.Boot.Retries {
			err := poller.Start()
			if err == nil {
				break
			}
			r++
			log.WithError(err).WithField("retries_left", p.cfg.Boot.Retries-r).Error("error fetching runways (initial)")
			if r == p.cfg.Boot.Retries {
				log.Fatal("error fetching runways (initially), no retries left")
			}
			time.Sleep(p.cfg.Boot.RetryCooldown)
		}
		defer poller.Stop()

		rawChan = psub.Updates()
	} else {
		data, err := ioutil.ReadFile(p.cfg.URL)
		if err != nil {
			log.WithError(err).WithField("filename", p.cfg.URL).Fatal("error loading file")
		}
		ch := make(chan []byte, 1)
		ch <- data
		rawChan = ch
	}

loop:
	for {
		select {
		case raw := <-rawChan:
			log.Debug("got update from vatsim api poller")
			p.parseRunways(raw)
		case <-p.stop:
			break loop
		}
	}
}

func (p *Provider) parseRunways(data []byte) {
	l := log.WithField("func", "parseRunways")

	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		r1, r2, err := parseRunway(line)
		if err != nil {
			l.WithError(err).Debug("error parsing runway")
			continue
		}

		p.dataLock.Lock()
		icao := r1.ICAO
		rwmap, found := p.runways[icao]
		if !found {
			rwmap = map[string]*Runway{}
			p.runways[icao] = rwmap
		}

		if ex, found := rwmap[r1.Ident]; !found || *ex != *r1 {
			rwmap[r1.Ident] = r1
			update := pubsub.Update{
				UType: pubsub.UpdateTypeSet,
				OType: ObjecTypeRunway,
				Obj:   *r1,
			}
			p.Notify(update)
		}

		if ex, found := rwmap[r2.Ident]; !found || *ex != *r2 {
			rwmap[r2.Ident] = r2
			update := pubsub.Update{
				UType: pubsub.UpdateTypeSet,
				OType: ObjecTypeRunway,
				Obj:   *r2,
			}
			p.Notify(update)
		}

		p.dataLock.Unlock()
	}
	p.SetDataReady(true)
}
