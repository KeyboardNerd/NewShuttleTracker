package YAST

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/remind101/migrate"
)

// PgSQL postgresql database implementation of Database Interface
// using cache for quick response to API request
type PgSQL struct {
	Url             string
	DB              *sql.DB
	CachedLatestLog map[string]*ShuttleLog  // vehicle id -> shuttle log
	CachedRoute     map[string]*ClosedRoute // route id -> closed route
}

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

const (
	insertMapPoint = `INSERT INTO map_point (longitude, latitude, angle, speed) VALUES ($1, $2, $3, $4) RETURNING id`
	// select or insert the shuttle's meta data if the shuttle is not found
	soiShuttleMeta = `WITH new_shuttle_meta AS ( 
							INSERT INTO shuttle_meta (remote_shuttle_id, shuttle_name) 
							SELECT CAST($1 AS VARCHAR), CAST($2 AS VARCHAR) 
							WHERE NOT EXISTS (SELECT remote_shuttle_id FROM shuttle_meta WHERE remote_shuttle_id=$1)
							RETURNING id)
						SELECT id FROM shuttle_meta WHERE remote_shuttle_id = $1
						UNION
						SELECT id FROM new_shuttle_meta`
	insertShuttleLog = `INSERT INTO shuttle_log (map_point_id, shuttle_meta_id, created_at) VALUES($1, $2, CURRENT_TIMESTAMP)`
	selectShuttleLog = ` 
					SELECT shuttle_log.id, shuttle_name, status, created_at, longitude, latitude, angle, speed
						FROM shuttle_log 
							LEFT JOIN shuttle_meta ON shuttle_log.shuttle_meta_id = shuttle_meta.id
							LEFT JOIN map_point ON shuttle_log.map_point_id = map_point.id
                     WHERE shuttle_meta.remote_shuttle_id = $1
						`
	insertRoutePath = `
		INSERT INTO route_path (route_id, map_point_id, ordering) VALUES ($1, $2, $3)
	`
	insertRouteInstance = `
		INSERT INTO route (name) VALUES ($1) RETURNING id
	`
	selectRoute = `
		SELECT route.id, longitude, latitude, angle, speed
		FROM map_point
		JOIN route_path ON route_path.map_point_id = map_point.id
		JOIN route ON route.id = route_path.route_id
		WHERE route.name = $1
		ORDER BY route_path.ordering
	`
	selectRouteMeta = `
		SELECT id, route_name FROM route WHERE name = $1
	`
)

// Open the database connection and initialize caches
func (pg *PgSQL) Open() {
	db, err := sql.Open("postgres", pg.Url)
	if err != nil {
		panic("Failed to connect to database")
	}
	pg.DB = db
	fmt.Printf("Started database migration\n")
	err = migrate.Exec(db, migrate.Up, migrations...)
	if err != nil {
		panic("Data migration failed\n")
	}
	fmt.Printf("Finished database migration\n")
	pg.CachedLatestLog = make(map[string]*ShuttleLog)
	pg.CachedRoute = make(map[string]*ClosedRoute)
}

// InsertClosedRoute inserts route into database and return the route with database ID and error
func (pg *PgSQL) InsertClosedRoute(route *ClosedRoute) error {
	tx, err := pg.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	// insert route meta data
	err = tx.QueryRow(insertRouteInstance, route.Name).Scan(&route.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// insert the map points
	for i, v := range route.RoutePoints {
		err = tx.QueryRow(insertMapPoint, v.X, v.Y, v.Angle, v.Speed).Scan(&v.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		// insert the path point
		_, err = tx.Exec(insertRoutePath, route.ID, v.ID, i)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

// SelectClosedRoute selects route by its external routeName from cache first, if it's missing, select from the database
func (pg *PgSQL) SelectClosedRoute(routeName string) (*ClosedRoute, error) {
	// if a shuttle id is missing in the cache, then query the database
	if r, ok := pg.CachedRoute[routeName]; ok {
		return r, nil
	}
	// query database
	tx, err := pg.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	vectors := []*Vector{}
	rows, err := tx.Query(selectRoute, routeName)
	if err != nil {
		return nil, err
	}
	var internalID int
	for rows.Next() {
		v := &Vector{}
		err = rows.Scan(&internalID, &v.X, &v.Y, &v.Angle, &v.Speed)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, v)
	}
	route := &ClosedRoute{RoutePoints: vectors, Name: routeName}
	tx.QueryRow(selectRouteMeta, routeName).Scan(&route.Name)
	pg.CachedRoute[routeName] = route
	return route, nil
}

func (pg *PgSQL) InsertStop(stop *Stop) error {
	panic("shit")
}

func (pg *PgSQL) SelectStop(stopId string) (*Stop, error) {
	panic("shit")
}

func (pg *PgSQL) SelectStopOnRoute(stopId string) ([]*Stop, error) {
	panic("shit")
}

// InsertShuttleLog to database
func (pg *PgSQL) InsertShuttleLog(log *ShuttleLog) error {
	tx, err := pg.DB.Begin()

	if err != nil {
		return err
	}
	defer tx.Commit()
	err = tx.QueryRow(insertMapPoint, log.Location.X, log.Location.Y, log.Location.Angle, log.Location.Speed).Scan(&log.Location.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	var (
		shuttle_meta_id sql.NullInt64
		shuttleName     sql.NullString
	)
	err = shuttleName.Scan(log.Name)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.QueryRow(soiShuttleMeta, log.VehicleID, shuttleName).Scan(&shuttle_meta_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(insertShuttleLog, log.Location.ID, shuttle_meta_id)
	if err != nil {
		tx.Rollback()
		return err
	}
	pg.CachedLatestLog[log.VehicleID] = log
	return nil
}

// SelectShuttleLog selects all shuttle logs of a shuttle specified by its remote id
func (pg *PgSQL) SelectShuttleLog(remoteShuttleID string) ([]*ShuttleLog, error) {
	var logs []*ShuttleLog
	v := &Vector{}
	s := &ShuttleLog{Location: v}
	tx, err := pg.DB.Begin()

	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	rows, err := tx.Query(selectShuttleLog, remoteShuttleID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		name := sql.NullString{}
		status := sql.NullString{}
		rows.Scan(&s.ID, &name, &status, &s.CreatedAt, &v.X, &v.Y, &v.Angle, &v.Speed)
		if name.Valid {
			s.Name = name.String
		}
		if status.Valid {
			s.Status = status.String
		}
		logs = append(logs, s)
	}
	return logs, nil
}

// SelectLatestLog fetches the latest shuttle's log from database
func (pg *PgSQL) SelectLatestLog(logid string) (*ShuttleLog, error) {
	if v, ok := pg.CachedLatestLog[logid]; ok {
		return v, nil
	}
	// query database if not in cache ( lazy load requires a special table to store the latest time stamp )
	return nil, fmt.Errorf("Shuttle Log with Vehicle ID '%s' not found in cache", logid)
}

// Close connection to database and clean caches
func (pg *PgSQL) Close() {
	pg.DB.Close()
	pg.CachedLatestLog = nil
	pg.CachedRoute = nil
}
