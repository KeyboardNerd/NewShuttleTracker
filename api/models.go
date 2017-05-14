package api

import (
	"encoding/json"

	"github.com/keyboardnerd/yastserver/database"
)

type ResStat struct {
	Response string `json:"_stat"`
	Info     string `json:"_info"`
}

type Context struct {
	DB database.Database
}

type ApiVector struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
	Speed float64 `json:"speed"`
}

type ApiShuttleLog struct {
	ResStat

	VehicleID string    `json:"id"`
	Location  ApiVector `json:"location"`
	Status    string    `json:"stat"`
}

type ApiClosedRoute struct {
	ResStat

	Locations []ApiVector `json:"location"`
	Name      string      `json:"name"`
}

func Stat(status, information string) []byte {
	em, err := json.Marshal(ResStat{status, information})
	if err != nil {
		panic("server broken")
	}
	return em
}

func (ar *ApiClosedRoute) FromDatabase(p *database.ClosedRoute) error {
	for _, r := range p.RoutePoints {
		av := ApiVector{}
		av.FromDatabase(r)
		ar.Locations = append(ar.Locations, av)
	}
	ar.Name = p.Name
	return nil
}

func (ar *ApiClosedRoute) ToDatabase() (*database.ClosedRoute, error) {
	r := &database.ClosedRoute{}
	for _, loc := range ar.Locations {
		v, err := loc.ToDatabase()
		if err != nil {
			return nil, err
		}
		r.RoutePoints = append(r.RoutePoints, v)
	}
	r.Name = ar.Name
	return r, nil
}

func (av *ApiVector) FromDatabase(p *database.Vector) error {
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return nil
}

func (p *ApiVector) ToDatabase() (*database.Vector, error) {
	av := &database.Vector{}
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return av, nil
}

func (alog *ApiShuttleLog) FromDatabase(log *database.ShuttleLog) error {
	alog.VehicleID = log.VehicleID
	alog.Status = log.Status
	av := ApiVector{}
	av.FromDatabase(log.Location)
	alog.Location = av
	return nil
}
