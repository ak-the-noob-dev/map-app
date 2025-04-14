package tileserver

// Street represents a road or street with its name and location
type Street struct {
	Name    string  `json:"name"`
	Highway string  `json:"highway"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// Area represents a named area or neighborhood
type Area struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Important bool    `json:"important"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

// POI represents a point of interest
type POI struct {
	Name      string  `json:"name"`
	Amenity   string  `json:"amenity"`
	Shop      string  `json:"shop"`
	Leisure   string  `json:"leisure"`
	Important bool    `json:"important"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

// Features contains all features for the current view
type Features struct {
	Streets []Street `json:"streets"`
	Areas   []Area   `json:"areas"`
	POIs    []POI    `json:"pois"`
}
