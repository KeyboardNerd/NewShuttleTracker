package api

import (
	"encoding/json"

	"github.com/keyboardnerd/yastserver/database"
)

type ResStat struct {
	Response string `json:"_stat"`
	Info     string `json:"_info"`
}

type srvContext struct {
	DB database.Database
}

type ApiMapPoint struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
	Speed float64 `json:"speed"`
}

type ApiShuttleLog struct {
	ResStat

	VehicleID string      `json:"id"`
	Location  ApiMapPoint `json:"location"`
	Status    string      `json:"stat"`
}

type ApiRoute struct {
	ResStat

	Locations []ApiMapPoint `json:"location"`
	Name      string        `json:"name"`
}

func Stat(status, information string) []byte {
	em, err := json.Marshal(ResStat{status, information})
	if err != nil {
		panic("server broken")
	}
	return em
}

func (ar *ApiRoute) FromDatabase(p *database.Route) error {
	for _, r := range p.RoutePoints {
		av := ApiMapPoint{}
		av.FromDatabase(r)
		ar.Locations = append(ar.Locations, av)
	}
	ar.Name = p.Name
	return nil
}

func (ar *ApiRoute) ToDatabase() (*database.Route, error) {
	r := &database.Route{}
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

func (av *ApiMapPoint) FromDatabase(p *database.MapPoint) error {
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return nil
}

func (p *ApiMapPoint) ToDatabase() (*database.MapPoint, error) {
	av := &database.MapPoint{}
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return av, nil
}

func (alog *ApiShuttleLog) FromDatabase(log *database.ShuttleLog) error {
	alog.VehicleID = log.VehicleID
	alog.Status = log.Status
	av := ApiMapPoint{}
	av.FromDatabase(log.Location)
	alog.Location = av
	return nil
}
