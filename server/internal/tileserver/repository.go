package tileserver

import (
	"server/internal/database"
	"strings"

	"go.uber.org/zap"
)

type Repository interface {
	GetTile(z, x, y int) ([]byte, error)
	GetFeatures(bbox []string, zoom int) (*Features, error)
	SearchLocation(query string) ([]map[string]interface{}, error)
	GetMapInfo() (map[string]interface{}, error)
}

type tileRepository struct {
	db database.Service
	logger *zap.Logger
}

func NewTileRepository(db database.Service,logger *zap.Logger) Repository {
	return &tileRepository{
		db: db,
		logger: logger,
	}
}

// -------------------- TILE --------------------

func (r *tileRepository) GetTile(z, x, y int) ([]byte, error) {
	query := `
	SELECT ST_AsMVT(q, 'osm_layer', 4096, 'geom') AS tile
	FROM (
		-- Major roads
		SELECT 
			osm_id, 
			name, 
			highway, 
			ref,
			oneway,
			surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_roads
		WHERE way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- All roads and paths
		SELECT 
			osm_id, 
			name, 
			highway, 
			ref,
			oneway,
			surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_line
		WHERE highway IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Water features (rivers, streams, etc.)
		SELECT 
			osm_id, 
			name, 
			waterway, 
			NULL AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_line
		WHERE waterway IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Railways
		SELECT 
			osm_id, 
			name, 
			railway, 
			NULL AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_line
		WHERE railway IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Land use areas (parks, forests, residential areas)
		SELECT 
			osm_id, 
			name, 
			landuse, 
			NULL AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_polygon
		WHERE landuse IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Leisure areas (parks, playgrounds, sports)
		SELECT 
			osm_id, 
			name, 
			leisure, 
			NULL AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_polygon
		WHERE leisure IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Natural features (water bodies, woods)
		SELECT 
			osm_id, 
			name, 
			"natural", 
			NULL AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_polygon
		WHERE "natural" IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Buildings
		SELECT 
			osm_id, 
			name, 
			building, 
			"addr:housenumber" AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_polygon
		WHERE building IS NOT NULL AND way && ST_TileEnvelope($1, $2, $3)

		UNION ALL

		-- Administrative boundaries
		SELECT 
			osm_id, 
			name, 
			boundary, 
			admin_level AS ref,
			NULL AS oneway,
			NULL AS surface,
			ST_AsMVTGeom(way, ST_TileEnvelope($1, $2, $3), 4096, 256, true) AS geom
		FROM planet_osm_line
		WHERE boundary = 'administrative' AND way && ST_TileEnvelope($1, $2, $3)
	) AS q;`
	var tile []byte
	err := r.db.GetDB().QueryRow(query, z, x, y).Scan(&tile)
	return tile, err
}

// -------------------- FEATURES --------------------

func (r *tileRepository) GetFeatures(bbox []string, zoom int) (*Features, error) {
	features := &Features{
		Streets: []Street{},
		Areas:   []Area{},
		POIs:    []POI{},
	}

	if zoom >= 14 {
		query := `
		SELECT 
			name, 
			highway,
			ST_Y(ST_Transform(ST_LineInterpolatePoint(way, 0.5), 4326)) as lat,
			ST_X(ST_Transform(ST_LineInterpolatePoint(way, 0.5), 4326)) as lon
		FROM planet_osm_line
		WHERE 
			highway IS NOT NULL 
			AND name IS NOT NULL
			AND name != ''
			AND ST_Transform(way, 4326) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
		ORDER BY 
			CASE 
				WHEN highway = 'motorway' THEN 1
				WHEN highway = 'trunk' THEN 2
				WHEN highway = 'primary' THEN 3
				WHEN highway = 'secondary' THEN 4
				WHEN highway = 'tertiary' THEN 5
				ELSE 6
			END
		LIMIT 500`
		rows, err := r.db.GetDB().Query(query, bbox[0], bbox[1], bbox[2], bbox[3])
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s Street
				if err := rows.Scan(&s.Name, &s.Highway, &s.Lat, &s.Lon); err != nil {
					r.logger.Warn("Failed to scan street feature", zap.Error(err))
					continue
				}
				features.Streets = append(features.Streets, s)
			}
		}
	}

	if zoom >= 11 {
		query := `SELECT 
			name, 
			COALESCE(place, landuse, leisure, "natural") as type,
			CASE 
				WHEN place IN ('suburb', 'neighbourhood', 'district') OR "boundary" = 'administrative' THEN true
				ELSE false
			END as important,
			ST_Y(ST_Transform(ST_Centroid(way), 4326)) as lat,
			ST_X(ST_Transform(ST_Centroid(way), 4326)) as lon
		FROM planet_osm_polygon
		WHERE 
			name IS NOT NULL 
			AND name != ''
			AND (place IS NOT NULL OR landuse IS NOT NULL OR leisure = 'park' OR "natural" = 'water')
			AND ST_Transform(way, 4326) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
		LIMIT 100`
		rows, err := r.db.GetDB().Query(query, bbox[0], bbox[1], bbox[2], bbox[3])
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var a Area
				rows.Scan(&a.Name, &a.Type, &a.Important, &a.Lat, &a.Lon)
				features.Areas = append(features.Areas, a)
			}
		}
	}

	if zoom >= 16 {
		query := `SELECT 
			name, 
			amenity,
			shop,
			leisure,
			CASE 
				WHEN amenity IN ('hospital', 'school', 'university', 'police') 
					OR shop IN ('supermarket', 'mall') 
					OR tourism IN ('hotel', 'attraction') THEN true
				ELSE false
			END as important,
			ST_Y(ST_Transform(way, 4326)) as lat,
			ST_X(ST_Transform(way, 4326)) as lon
		FROM planet_osm_point
		WHERE 
			name IS NOT NULL 
			AND name != ''
			AND (amenity IS NOT NULL OR shop IS NOT NULL OR leisure IS NOT NULL OR tourism IS NOT NULL)
			AND ST_Transform(way, 4326) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
		LIMIT 100`
		rows, err := r.db.GetDB().Query(query, bbox[0], bbox[1], bbox[2], bbox[3])
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var p POI
				rows.Scan(&p.Name, &p.Amenity, &p.Shop, &p.Leisure, &p.Important, &p.Lat, &p.Lon)
				features.POIs = append(features.POIs, p)
			}
		}
	}

	return features, nil
}

// -------------------- SEARCH --------------------

func (r *tileRepository) SearchLocation(query string) ([]map[string]interface{}, error) {
	searchQuery := `
	WITH combined_results AS (
		-- Search in points (amenities, POIs)
		SELECT 
			name, 
			amenity AS type,
			ST_Y(ST_Transform(way, 4326)) AS lat,
			ST_X(ST_Transform(way, 4326)) AS lon,
			1 AS priority
		FROM planet_osm_point
		WHERE 
			name IS NOT NULL 
			AND LOWER(name) LIKE $1
			AND amenity IS NOT NULL
			
		UNION ALL
		
		-- Search in roads
		SELECT 
			name, 
			highway AS type,
			ST_Y(ST_Transform(ST_LineInterpolatePoint(way, 0.5), 4326)) AS lat,
			ST_X(ST_Transform(ST_LineInterpolatePoint(way, 0.5), 4326)) AS lon,
			2 AS priority
		FROM planet_osm_line
		WHERE 
			name IS NOT NULL 
			AND LOWER(name) LIKE $1
			AND highway IS NOT NULL
			
		UNION ALL
		
		-- Search in polygons (buildings, areas)
		SELECT 
			name, 
			COALESCE(building, landuse, leisure, "natural") AS type,
			ST_Y(ST_Transform(ST_Centroid(way), 4326)) AS lat,
			ST_X(ST_Transform(ST_Centroid(way), 4326)) AS lon,
			3 AS priority
		FROM planet_osm_polygon
		WHERE 
			name IS NOT NULL 
			AND LOWER(name) LIKE $1
	)
	SELECT name, type, lat, lon
	FROM combined_results
	ORDER BY 
		priority,
		CASE WHEN LOWER(name) = LOWER($2) THEN 0 ELSE 1 END,
		LENGTH(name)
	LIMIT 10
	`

	// rows, err := r.db.GetDB().Query(searchQuery, "%"+strings.ToLower(query)+"%", query)
	safeQuery := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.GetDB().Query(searchQuery, safeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var name, typ string
		var lat, lon float64
		rows.Scan(&name, &typ, &lat, &lon)
		results = append(results, map[string]interface{}{
			"name": name,
			"type": typ,
			"lat":  lat,
			"lon":  lon,
		})
	}

	return results, nil
}

// -------------------- MAP INFO --------------------

func (r *tileRepository) GetMapInfo() (map[string]interface{}, error) {
	query := `
	SELECT 
		(SELECT COUNT(*) FROM planet_osm_roads),
		(SELECT COUNT(*) FROM planet_osm_polygon WHERE building IS NOT NULL),
		(SELECT COUNT(*) FROM planet_osm_point WHERE amenity IS NOT NULL OR shop IS NOT NULL),
		(SELECT MIN(ST_X(ST_Transform(ST_Centroid(way), 4326))) FROM planet_osm_polygon),
		(SELECT MIN(ST_Y(ST_Transform(ST_Centroid(way), 4326))) FROM planet_osm_polygon),
		(SELECT MAX(ST_X(ST_Transform(ST_Centroid(way), 4326))) FROM planet_osm_polygon),
		(SELECT MAX(ST_Y(ST_Transform(ST_Centroid(way), 4326))) FROM planet_osm_polygon)
	`

	var roadCount, buildingCount, poiCount int
	var minLon, minLat, maxLon, maxLat float64

	err := r.db.GetDB().QueryRow(query).Scan(&roadCount, &buildingCount, &poiCount, &minLon, &minLat, &maxLon, &maxLat)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"statistics": map[string]int{
			"roads":              roadCount,
			"buildings":          buildingCount,
			"points_of_interest": poiCount,
		},
		"bounds": map[string]float64{
			"min_lat": minLat,
			"min_lon": minLon,
			"max_lat": maxLat,
			"max_lon": maxLon,
		},
		"version":      "1.0.0",
		"last_updated": "2025-04-03",
	}, nil
}
