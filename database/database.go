package database

import "time"

type Database interface {
	// Initialize
	Open()
	// Insert a shuttle log into database
	InsertShuttleLog(*ShuttleLog) error
	// return the latest log of a shuttle by shuttle name
	SelectLatestLog(string) (*ShuttleLog, error)
	// Insert a closed route to database
	InsertRoute(*Route) error
	// Select a closed route to database by route name
	SelectRoute(string) (*Route, error)
	// Insert a stop to database
	InsertStop(*Stop) error
	// Select a stop from database by stop name
	SelectStop(string) (*Stop, error)
	// Select all stops on a route by route name
	SelectStopOnRoute(string) ([]*Stop, error)
	// close
	Close()
}

type Model struct {
	ID int64
}

type MapPoint struct {
	Model

	X     float64
	Y     float64
	Angle float64
	Speed float64
}

type Shuttle struct {
	Model

	Name       string
	RemoteName string
}

// ShuttleLog stores the logging information of the shuttle
// directly mapped to remote packet
type ShuttleLog struct {
	Model

	VehicleID string
	Status    string
	Location  *MapPoint
	Name      string
	CreatedAt time.Time
}

// Route contains a list of vectors in the database with well defined ordering
// Route should be a closed loop with start
type Route struct {
	Model

	RoutePoints []*MapPoint
	Name        string
}

// Stop represents a vector on a route
type Stop struct {
	Model

	Location *MapPoint
	Route    *Route
	StopID   string
	Name     string
}
