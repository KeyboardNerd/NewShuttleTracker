package YAST

import (
	"errors"
	"fmt"
	"sync"
)

// MockDatabase for testing
type MockDatabase struct {
	sync.Mutex

	LogTabel    []ShuttleLog           // ( mock main database table)
	LatestTabel map[string]*ShuttleLog // contains reference to logtabel ( mock foreign key )
	RouteTabel  map[string]*ClosedRoute
	RouteID     int
}

func (db *MockDatabase) Open() {
	db.Lock()
	defer db.Unlock()
	db.LogTabel = make([]ShuttleLog, 100)
	db.LatestTabel = make(map[string]*ShuttleLog)
	db.RouteID = 0
}

func (db *MockDatabase) InsertShuttleLog(log *ShuttleLog) error {
	db.Lock()
	defer db.Unlock()
	fmt.Printf("Insert shuttle log %#v\n", log)
	db.LogTabel = append(db.LogTabel, *log)
	db.LatestTabel[log.VehicleID] = &db.LogTabel[len(db.LogTabel)-1]

	return nil
}

func (db *MockDatabase) SelectLatestLog(vid string) (*ShuttleLog, error) {
	db.Lock()
	defer db.Unlock()
	if log, ok := db.LatestTabel[vid]; ok {
		return log, nil
	}

	return nil, fmt.Errorf("vehicle key (%s) not found in database\n", vid)
}

func (db *MockDatabase) InsertClosedRoute(route *ClosedRoute) error {
	db.Lock()
	defer db.Unlock()
	db.RouteTabel[string(db.RouteID)] = route
	db.RouteID++
	return nil
}

func (db *MockDatabase) SelectClosedRoute(rid string) (*ClosedRoute, error) {
	db.Lock()
	defer db.Unlock()
	if v, ok := db.RouteTabel[rid]; ok {
		return v, nil
	}
	return nil, errors.New("route not found")
}

func (db *MockDatabase) Close() {
	db.Lock()
	defer db.Unlock()
	db.LogTabel = nil
	db.LatestTabel = nil
	db.RouteID = 0
	db.RouteTabel = nil
}

// Insert a stop to database
func (db *MockDatabase) InsertStop(*Stop) error {
	panic("The thing is not implemented ")
	return nil
}

// Select a stop from database by stop name
func (db *MockDatabase) SelectStop(string) (*Stop, error) {
	panic("The thing is not implemented ")
	return nil, nil
}

// Select all stops on a route by route name
func (db *MockDatabase) SelectStopOnRoute(string) ([]*Stop, error) {
	panic("The thing is not implemented ")
	return nil, nil
}
