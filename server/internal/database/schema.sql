-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create a table for streets
CREATE TABLE IF NOT EXISTS streets (
    id SERIAL PRIMARY KEY,
    name TEXT,
    type TEXT,
    geom geometry(LineString, 4326)
);

-- Insert some sample data
INSERT INTO streets (name, type, geom) VALUES
    ('Sample Street 1', 'road', ST_GeomFromText('LINESTRING(0 0, 1 1)', 4326)),
    ('Sample Street 2', 'road', ST_GeomFromText('LINESTRING(1 1, 2 2)', 4326)); 