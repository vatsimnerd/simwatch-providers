package ourairports

// Reference:
// "id","airport_ref","airport_ident","length_ft","width_ft","surface","lighted","closed","le_ident","le_latitude_deg","le_longitude_deg","le_elevation_ft","le_heading_degT","le_displaced_threshold_ft","he_ident","he_latitude_deg","he_longitude_deg","he_elevation_ft","he_heading_degT","he_displaced_threshold_ft"
// 239399,2434,"EGLL",12799,164,"ASP",1,0,"09L",51.4775,-0.489428,79,89.6,1007,"27R",51.4777,-0.433264,78,269.6,

type Runway struct {
	ICAO        string  `json:"icao"`
	LengthFt    int     `json:"length_ft"`
	WidthFt     int     `json:"width_ft"`
	Surface     string  `json:"surface"`
	Lighted     bool    `json:"lighted"`
	Closed      bool    `json:"closed"`
	Ident       string  `json:"ident"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	ElevationFt int     `json:"elev_ft"`
	Heading     float64 `json:"hdg"`
	ActiveTO    bool    `json:"active_to"`
	ActiveLnd   bool    `json:"active_lnd"`
}
