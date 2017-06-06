package database

import "github.com/remind101/migrate"

var migrations = []migrate.Migration{
	{
		// each shuttle and stop belong to a route
		ID: 1,
		Up: migrate.Queries([]string{
			`CREATE TABLE IF NOT EXISTS shuttle(
					id SERIAL PRIMARY KEY,
					remote_name VARCHAR(64) UNIQUE NOT NULL,
					CONSTRAINT remote_name CHECK(char_length(remote_name) > 0),
					name VARCHAR(64) UNIQUE NOT NULL
				)`,
			`CREATE INDEX ON shuttle(name)`,
			`CREATE TABLE IF NOT EXISTS route(
					id SERIAL PRIMARY KEY,
					name VARCHAR(64) UNIQUE NOT NULL,
					CONSTRAINT name CHECK(char_length(name) > 0)
				)`,
			`CREATE INDEX ON route(name)`,
			`CREATE TABLE IF NOT EXISTS shuttle_point(
					id SERIAL PRIMARY KEY,
					longitude FLOAT,
					latitude FLOAT,
					angle FLOAT,
					speed FLOAT,
					status VARCHAR(64),
					shuttle_id INT REFERENCES shuttle(id) ON DELETE CASCADE,
					created_at TIMESTAMP WITH TIME ZONE
				)`,

			`CREATE TABLE IF NOT EXISTS route_point(
					id SERIAL PRIMARY KEY,
					route_id INT REFERENCES route(id) ON DELETE CASCADE,
					ordering INT NOT NULL,
					longitude FLOAT,
					latitude FLOAT,
					angle FLOAT,
					speed FLOAT
				)`,
			`CREATE INDEX ON route_point(route_id)`,
			`CREATE TABLE IF NOT EXISTS stop(
					id SERIAL PRIMARY KEY,
					name VARCHAR(64) UNIQUE NOT NULL,
					route_id INT REFERENCES route(id) ON DELETE CASCADE,
					route_point INT REFERENCES route_point(id) ON DELETE CASCADE
				)`,
		}),
		Down: migrate.Queries([]string{
			`DROP TABLE IF EXISTS shuttle_log, shuttle_meta, route, route_path, stop, stop_meta, shuttle_map_point`,
		}),
	},
}
