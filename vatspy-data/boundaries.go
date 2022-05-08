package vatspydata

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	log "github.com/sirupsen/logrus"
	"github.com/vatsimnerd/util/mapupdate"
)

func (p *Provider) parseBoundaries(raw []byte) error {
	log.Info("parsing boundaries")
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	if err != nil {
		return err
	}

	bdrs := make(map[string]Boundaries)
	for _, feat := range fc.Features {
		// All objects are multipolygons as in April 2022
		if mp, ok := feat.Geometry.(orb.MultiPolygon); ok {
			var b Boundaries
			b.Points = make([][]Point, 0)

			for _, poly := range mp {
				// preallocate 128 elements
				points := make([]Point, 0, 128)
				for _, coords := range poly[0] {
					// This is a final []float64 array representing a coordinate
					// thus must be of size 2
					if len(coords) != 2 {
						continue
					}
					points = append(points, Point{Lng: coords[0], Lat: coords[1]})
				}
				b.Points = append(b.Points, points)
			}
			bound := mp.Bound()
			center := bound.Center()
			min := bound.Min
			max := bound.Max
			b.ID = readStringProp(feat, "id")
			b.Center = Point{Lat: center.Lat(), Lng: center.Lon()}
			b.Min = Point{Lat: min.Lat(), Lng: min.Lon()}
			b.Max = Point{Lat: max.Lat(), Lng: max.Lon()}

			b.Region = readStringProp(feat, "region")
			b.IsOceanic = readStringProp(feat, "oceanic") == "1"
			b.Division = readStringProp(feat, "division")
			bdrs[b.ID] = b
		}
	}

	// todo proper update, i.e. update firs
	_, _ = mapupdate.Update[Boundaries, mapupdate.Comparable[Boundaries]](p.bdrs, bdrs, &p.dataLock)

	return nil
}

func readStringProp(feat *geojson.Feature, key string) string {
	if value, found := feat.Properties[key]; found {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
