import React, { useEffect, useRef, useState } from 'react'
import maplibregl, { Map, Marker, Popup } from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

interface MapViewProps {}

const MapView: React.FC<MapViewProps> = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<Map | null>(null)
  const labelMarkers = useRef<Marker[]>([])
  const [activePopup, setActivePopup] = useState<Popup | null>(null)

  // Emoji-based POI icon
  const getPOIIcon = (props: any): string => {
    const { amenity, shop, leisure } = props
    if (amenity === 'hospital' || amenity === 'clinic') return '🏥'
    if (amenity === 'school') return '🏫'
    if (amenity === 'restaurant') return '🍴'
    if (shop === 'supermarket') return '🛒'
    if (leisure === 'park') return '🌳'
    if (amenity === 'place_of_worship') return '⛪'
    return '📍'
  }

  const addLabelMarker = (map: Map, coords: [number, number], html: string, className: string) => {
    const el = document.createElement('div')
    el.className = className
    el.innerHTML = html
    const marker = new Marker({ element: el, interactive: false }).setLngLat(coords).addTo(map)
    labelMarkers.current.push(marker)
  }

  const clearLabels = () => {
    labelMarkers.current.forEach((m) => m.remove())
    labelMarkers.current = []
  }

  const loadLabels = async () => {
    if (!mapRef.current) return
    const map = mapRef.current
    const bounds = map.getBounds()
    const zoom = map.getZoom()

    const bbox = `${bounds.getWest()},${bounds.getSouth()},${bounds.getEast()},${bounds.getNorth()}`
    try {
      const res = await fetch(`http://localhost:8080/api/features?bbox=${bbox}&zoom=${zoom}`)
      const data = await res.json()

      clearLabels()

      // Roads
      if (zoom >= 14) {
        data.streets.forEach((street: any) => {
          if (zoom < 16 && !['motorway', 'trunk', 'primary', 'secondary', 'tertiary'].includes(street.highway)) return
          if (street.name) {
            addLabelMarker(map, [street.lon, street.lat], street.name, 'street-label')
          }
        })
      }

      // Areas
      if (zoom >= 11) {
        data.areas.forEach((area: any) => {
          if (zoom < 13 && !area.important) return
          if (area.name) {
            const fontSize = area.important ? 16 : 14
            addLabelMarker(
              map,
              [area.lon, area.lat],
              `<div style="font-size:${fontSize}px;">${area.name}</div>`,
              'area-label',
            )
          }
        })
      }

      // POIs
      if (zoom >= 16) {
        data.pois.forEach((poi: any) => {
          const icon = getPOIIcon(poi)
          if (poi.name) {
            // Emoji icon marker
            addLabelMarker(map, [poi.lon, poi.lat], icon, 'poi-icon')
            // Label below if important
            if (poi.important) {
              addLabelMarker(map, [poi.lon, poi.lat - 0.00008], poi.name, 'poi-label')
            }
          }
        })
      }
    } catch (err) {
      console.error('Label load error', err)
    }
  }

  useEffect(() => {
    if (!mapContainerRef.current) return

    const map = new maplibregl.Map({
      container: mapContainerRef.current,
      style: {
        version: 8,
        sources: {
          vectortiles: {
            type: 'vector',
            tiles: ['http://localhost:8080/tiles/{z}/{x}/{y}.mvt'],
            minzoom: 0,
            maxzoom: 18,
          },
        },
        layers: [
          {
            id: 'background',
            type: 'background',
            paint: {
              'background-color': '#f8f4f0',
            },
          },
          {
            id: 'water',
            type: 'fill',
            source: 'vectortiles',
            'source-layer': 'water',
            paint: {
              'fill-color': '#aad3df',
            },
          },
          {
            id: 'buildings',
            type: 'fill',
            source: 'vectortiles',
            'source-layer': 'building',
            paint: {
              'fill-color': '#d4cbc8',
              'fill-opacity': 0.9,
            },
          },
          {
            id: 'highways',
            type: 'line',
            source: 'vectortiles',
            'source-layer': 'transportation',
            filter: ['==', 'class', 'motorway'],
            paint: {
              'line-color': '#e892a2',
              'line-width': ['interpolate', ['linear'], ['zoom'], 10, 2, 16, 6],
            },
          },
          {
            id: 'primary-roads',
            type: 'line',
            source: 'vectortiles',
            'source-layer': 'transportation',
            filter: ['in', 'class', 'primary', 'secondary', 'tertiary'],
            paint: {
              'line-color': '#ffffff',
              'line-width': ['interpolate', ['linear'], ['zoom'], 10, 1, 16, 5],
            },
          },
        ],
      },
      center: [80.2707, 13.0827],
      zoom: 13,
    })

    mapRef.current = map

    map.on('load', () => loadLabels())
    map.on('moveend', () => loadLabels())
    map.on('zoomend', () => loadLabels())

    return () => {
      map.remove()
    }
  }, [])

  return (
    <>
      <div ref={mapContainerRef} style={{ width: '100%', height: '100vh', position: 'relative' }} />
      <input
        type="text"
        className="search-input"
        placeholder="Search..."
        onKeyDown={async (e) => {
          if (e.key === 'Enter') {
            const value = (e.target as HTMLInputElement).value.trim()
            if (!value) return
            try {
              const res = await fetch(`http://localhost:8080/api/search?q=${encodeURIComponent(value)}`)
              const results = await res.json()
              if (results.length > 0) {
                const first = results[0]
                mapRef.current?.setView([first.lon, first.lat], 17)
                new Marker().setLngLat([first.lon, first.lat]).addTo(mapRef.current!)
              } else {
                alert('No results found')
              }
            } catch {
              alert('Search failed')
            }
          }
        }}
        style={{
          position: 'absolute',
          top: 10,
          left: 10,
          zIndex: 999,
          padding: '6px 10px',
          borderRadius: '4px',
          border: '1px solid #ccc',
        }}
      />
    </>
  )
}

export default MapView
