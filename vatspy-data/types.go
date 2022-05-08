package vatspydata

type (
	// Point is a map point
	Point struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	// Boundaries object
	Boundaries struct {
		ID        string    `json:"id"`
		IsOceanic bool      `json:"is_oceanic"`
		Region    string    `json:"region"`
		Division  string    `json:"division"`
		Min       Point     `json:"min"`
		Max       Point     `json:"max"`
		Center    Point     `json:"center"`
		Points    [][]Point `json:"points"`
	}

	// Country object
	Country struct {
		Name              string `json:"name"`
		Prefix            string `json:"prefix"`
		ControlCustomName string `json:"control_custom_name"`
	}

	// Airport object
	AirportMeta struct {
		ICAO     string `json:"icao"`
		Name     string `json:"name"`
		Position Point  `json:"position"`
		IATA     string `json:"iata"`
		FIRID    string `json:"fir_id"`
		IsPseudo bool   `json:"is_pseudo"`
	}

	// FIR object
	FIR struct {
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		Prefix     string     `json:"prefix"`
		ParentID   string     `json:"parent_id"`
		Boundaries Boundaries `json:"boundaries"`
	}

	// UIR object
	UIR struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		FIRIDs []string `json:"fir_ids"`
	}
)

func (c Country) NE(o Country) bool {
	return c != o
}

func (a AirportMeta) NE(o AirportMeta) bool {
	return a != o
}

func (b Boundaries) NE(o Boundaries) bool {
	if b.ID != o.ID ||
		b.IsOceanic != o.IsOceanic ||
		b.Region != o.Region ||
		b.Division != o.Division ||
		b.Min != o.Min ||
		b.Max != o.Max ||
		b.Center != o.Center {
		return true
	}

	if len(b.Points) != len(o.Points) {
		return true
	}

	for i := 0; i < len(b.Points); i++ {
		if len(b.Points[i]) != len(o.Points[i]) {
			for j := 0; j < len(b.Points[i]); j++ {
				if b.Points[i][j] != o.Points[i][j] {
					return true
				}
			}
		}
	}

	return false
}

func (f FIR) NE(o FIR) bool {
	return (f.ID != o.ID ||
		f.Name != o.Name ||
		f.Prefix != o.Prefix ||
		f.ParentID != o.ParentID || f.Boundaries.NE(o.Boundaries))
}

func (u UIR) NE(o UIR) bool {
	if u.ID != o.ID || u.Name != o.Name {
		return true
	}
	if len(u.FIRIDs) != len(o.FIRIDs) {
		return true
	}
	for i := 0; i < len(u.FIRIDs); i++ {
		if u.FIRIDs[i] != o.FIRIDs[i] {
			return true
		}
	}
	return false
}
