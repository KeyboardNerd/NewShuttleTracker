package YAST

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/remind101/migrate"
)

type PgSQL struct {
	Url             string
	DB              *sql.DB
	CachedLatestLog map[string]*ShuttleLog
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
			// route meta data
			`CREATE TABLE IF NOT EXISTS route_meta(
					id SERIAL PRIMARY KEY,
					route_name VARCHAR(64) NULL UNIQUE
				)`,
			// route
			`CREATE TABLE IF NOT EXISTS route(
					id SERIAL PRIMARY KEY,
					route_meta_id INT NULL REFERENCES route_meta(id) ON DELETE SET NULL
				)`,
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
					remote_shuttle_id VARCHAR(64) UNIQUE,
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
					stop_meta INT NULL REFERENCES route_meta(id) ON DELETE SET NULL
				)`,
			`CREATE INDEX ON stop(route_id)`,
		}),
		Down: migrate.Queries([]string{
			`DROP TABLE IF EXISTS shuttle_log, shuttle_meta, route, route_path, route_meta, stop, stop_meta, map_point;`,
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
)

func (pg *PgSQL) Open() {
	db, err := sql.Open("postgres", pg.Url)
	if err != nil {
		panic("Failed to connect to database")
	}
	pg.DB = db
	err = migrate.Exec(db, migrate.Up, migrations...)
	if err != nil {
		panic("Data migration failed")
	}
	pg.CachedLatestLog = make(map[string]*ShuttleLog)
}

func (pg *PgSQL) InsertClosedRoute(route *ClosedRoute) error {
	panic("shit")
}

func (pg *PgSQL) SelectClosedRoute(routeId string) (*ClosedRoute, error) {
	panic("shit")
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

func (pg *PgSQL) InsertShuttleLog(log *ShuttleLog) error {
	tx, err := pg.DB.Begin()

	if err != nil {
		return err
	}
	defer tx.Commit()
	err = tx.QueryRow(insertMapPoint, log.Location.X, log.Location.Y, log.Location.Angle, log.Location.Speed).Scan(&log.Location.ID)
	if err != nil {
		tx.Rollback()
		panic(err)
		return err
	}
	var (
		shuttle_meta_id sql.NullInt64
		shuttleName     sql.NullString
	)
	err = shuttleName.Scan(log.Name)
	if err != nil {
		tx.Rollback()
		panic(err)
		return err
	}
	err = tx.QueryRow(soiShuttleMeta, log.VehicleID, shuttleName).Scan(&shuttle_meta_id)
	fmt.Println(log.VehicleID)
	fmt.Println(shuttleName)
	if err != nil {
		tx.Rollback()
		panic(err)
		return err
	}

	_, err = tx.Exec(insertShuttleLog, log.Location.ID, shuttle_meta_id)
	if err != nil {
		tx.Rollback()
		fmt.Println(shuttle_meta_id)
		panic(err)
		return err
	}
	pg.CachedLatestLog[log.VehicleID] = log
	return nil
}

func (pg *PgSQL) SelectLog(logid string) (*ShuttleLog, error) {
	v := &Vector{}
	s := &ShuttleLog{Location: v}
	tx, err := pg.DB.Begin()

	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	name := sql.NullString{}
	status := sql.NullString{}
	err = tx.QueryRow(selectShuttleLog, logid).Scan(&s.ID, &name, &status, &s.CreatedAt, &v.X, &v.Y, &v.Angle, &v.Speed)
	if name.Valid {
		s.Name = name.String
	}
	if status.Valid {
		s.Status = status.String
	}
	if err != nil {
		return nil, err
	}
	fmt.Printf("Select Latest Log %v")
	return s, nil
}

func (pg *PgSQL) SelectLatestLog(logid string) (*ShuttleLog, error) {
	// TODO: initial value should be the latest one in database
	if v, ok := pg.CachedLatestLog[logid]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("Shuttle Log with Vehicle ID '%s' not found", logid)
}

func (pg *PgSQL) Close() {
	pg.DB.Close()
	pg.CachedLatestLog = nil
}
