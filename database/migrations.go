package database

import "github.com/remind101/migrate"

var migrations = []migrate.Migration{
	{
		ID: 1,
		Up: migrate.Queries([]string{
			// map point
			`CREATE TABLE IF NOT EXISTS map_point(
					id SERIAL PRIMARY KEY,
					longitude FLOAT,
					latitude FLOAT,
					angle FLOAT,
					speed FLOAT
				)`,
			// route
			`CREATE TABLE IF NOT EXISTS route(
					id SERIAL PRIMARY KEY,
					name VARCHAR(64) UNIQUE NOT NULL,
					CONSTRAINT name CHECK(char_length(name) > 0)
				)`,
			`CREATE INDEX ON route(name)`,
			// path of a route
			`CREATE TABLE IF NOT EXISTS route_path(
					id SERIAL PRIMARY KEY,
					route_id INT REFERENCES route(id) ON DELETE CASCADE,
					map_point_id INT REFERENCES map_point(id) ON DELETE CASCADE,
					ordering INT
				)`,
			`CREATE INDEX ON route_path(route_id)`,
			// shuttle meta data
			`CREATE TABLE IF NOT EXISTS shuttle_meta(
					id SERIAL PRIMARY KEY,
					remote_shuttle_id VARCHAR(64) UNIQUE NOT NULL,
					CONSTRAINT remote_shuttle_id CHECK(char_length(remote_shuttle_id) > 0),
					shuttle_name VARCHAR(64), 
					shuttle_route_id INT NULL REFERENCES route(id) ON DELETE SET NULL
				)`,
			// shuttle logs
			`CREATE TABLE IF NOT EXISTS shuttle_log(
					id SERIAL PRIMARY KEY,
					map_point_id INT REFERENCES map_point(id) ON DELETE CASCADE,
					shuttle_meta_id INT NULL REFERENCES shuttle_meta(id) ON DELETE SET NULL,
					status VARCHAR(64),
					created_at TIMESTAMP WITH TIME ZONE
				)`,
			// stop meta data
			`CREATE TABLE IF NOT EXISTS stop_meta(
					id SERIAL PRIMARY KEY,
					stop_name VARCHAR(64)
				)`,
			// stops
			`CREATE TABLE IF NOT EXISTS stop(
					id SERIAL PRIMARY KEY,
					route_id INT REFERENCES route(id) ON DELETE CASCADE,
					map_point_id INT REFERENCES map_point(id) ON DELETE CASCADE,
					stop_meta_id INT NULL REFERENCES stop_meta(id) ON DELETE SET NULL
				)`,
			`CREATE INDEX ON stop(route_id)`,
		}),
		Down: migrate.Queries([]string{
			`DROP TABLE IF EXISTS shuttle_log, shuttle_meta, route, route_path, stop, stop_meta, map_point`,
		}),
	},
}
