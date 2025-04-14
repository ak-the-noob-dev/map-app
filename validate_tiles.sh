#!/bin/bash

# Function to download and validate a tile
validate_tile() {
    local z=$1
    local x=$2
    local y=$3
    
    echo "Validating tile z=$z x=$x y=$y"
    
    # Download the tile
    curl -o "tile_${z}_${x}_${y}.mvt" "http://localhost:8080/tiles/${z}/${x}/${y}.mvt"
    
    # Check if the file was downloaded successfully
    if [ ! -f "tile_${z}_${x}_${y}.mvt" ]; then
        echo "Failed to download tile"
        return 1
    fi
    
    # Check file size
    local size=$(stat -f%z "tile_${z}_${x}_${y}.mvt")
    echo "Tile size: $size bytes"
    
    # Show file contents in hex
    echo "File contents (hex):"
    xxd -g 1 "tile_${z}_${x}_${y}.mvt"
    
    # Try to decode the tile
    echo "Decoded tile contents:"
    tippecanoe-decode "tile_${z}_${x}_${y}.mvt"
    
    # Clean up
    rm "tile_${z}_${x}_${y}.mvt"
    echo "----------------------------------------"
}

# Test a few different zoom levels and coordinates
validate_tile 0 0 0
validate_tile 1 0 0
validate_tile 1 1 0
validate_tile 2 1 1 