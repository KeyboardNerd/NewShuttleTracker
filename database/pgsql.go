package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/remind101/migrate"
)

// PgSQL postgresql database implementation of Database Interface
// using cache for quick response to API request
type PgSQL struct {
	URL             string
	DB              *sql.DB
	CachedLatestLog map[string]*ShuttleLog // vehicle id -> shuttle log
	CachedRoute     map[string]*Route      // route id -> closed route
}

// Open the database connection and initialize caches
func (pg *PgSQL) Open() {
	db, err := sql.Open("postgres", pg.URL)
	if err != nil {
		log.Fatalln("Failed to connect to database")
	}
	pg.DB = db
	log.Println("Started database migration")
	err = migrate.Exec(db, migrate.Up, migrations...)
	if err != nil {
		log.Fatalln("Data migration failed")
	}
	log.Println("Finished database migration")
	pg.CachedLatestLog = make(map[string]*ShuttleLog)
	pg.CachedRoute = make(map[string]*Route)
}

// ListRouteName gives a list of route names
func (pg *PgSQL) ListRouteName() ([]string, error) {
	tx, err := pg.DB.Begin()
	defer tx.Commit()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(selectAllRouteName)
	if err != nil {
		return nil, err
	}
	r := []string{}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		r = append(r, s)
	}
	return r, nil
}

// InsertRoute inserts route into database and return the route with database ID and error
func (pg *PgSQL) InsertRoute(route *Route) error {
	tx, err := pg.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	// insert route meta data
	err = tx.QueryRow(insertRoute, route.Name).Scan(&route.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// insert the map points
	for _, v := range route.RoutePoints {
		err = tx.QueryRow(insertRoutePath, v.X, v.Y, v.Angle, v.Speed).Scan(&v.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

// SelectRoute selects route by its external routeName from cache first, if it's missing, select from the database
func (pg *PgSQL) SelectRoute(routeName string) (*Route, error) {
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
	vectors := []*MapPoint{}
	route := &Route{Name: routeName}
	rows, err := tx.Query(selectRoute, routeName)
	if err != nil {
		return nil, err
	}
	var internalID int
	for rows.Next() {
		v := &MapPoint{}
		err = rows.Scan(&internalID, &v.X, &v.Y, &v.Angle, &v.Speed)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, v)
	}
	route.RoutePoints = vectors
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
	err = tx.QueryRow(insertShuttleLog, log.Location.X, log.Location.Y, log.Location.Angle, log.Location.Speed).Scan(&log.Location.ID)
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
	err = tx.QueryRow(soiShuttle, log.VehicleID, shuttleName).Scan(&shuttle_meta_id)
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
	v := &MapPoint{}
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
