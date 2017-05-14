package YAST

import "time"

type Database interface {
	// Initialize
	Open()
	// Insert a shuttle log into database
	InsertShuttleLog(*ShuttleLog) error
	// return the latest log of a shuttle by shuttle name
	SelectLatestLog(string) (*ShuttleLog, error)
	// Insert a closed route to database
	InsertClosedRoute(*ClosedRoute) error
	// Select a closed route to database by route name
	SelectClosedRoute(string) (*ClosedRoute, error)
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

type Vector struct {
	Model

	X     float64
	Y     float64
	Angle float64
	Speed float64
}

// ShuttleLog stores the logging information of the shuttle
// directly mapped to remote packet
type ShuttleLog struct {
	Model

	VehicleID string
	Status    string
	Location  *Vector
	Name      string
	CreatedAt time.Time
}

// ClosedRoute contains a list of vectors in the database with well defined ordering
// ClosedRoute should be a closed loop with start
type ClosedRoute struct {
	Model

	RoutePoints []*Vector
	// Start and End vectors should be a reference to vectors in RoutePoints
	Start int
	End   int

	Name string
}

// Stop represents a vector on a route
type Stop struct {
	Model

	Location *Vector
	Route    *ClosedRoute
	Name     string
}
