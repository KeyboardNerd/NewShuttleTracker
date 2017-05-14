package YAST

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	ERROR = "error"
	OK    = "ok"
)

type ResStat struct {
	Response string `json:"_stat"`
	Info     string `json:"_info"`
}

type Context struct {
	DB Database
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

func (ar *ApiClosedRoute) FromDatabase(p *ClosedRoute) error {
	for _, r := range p.RoutePoints {
		av := ApiVector{}
		av.FromDatabase(r)
		ar.Locations = append(ar.Locations, av)
	}
	ar.Name = p.Name
	return nil
}

func (ar *ApiClosedRoute) ToDatabase() (*ClosedRoute, error) {
	r := &ClosedRoute{}
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

func (av *ApiVector) FromDatabase(p *Vector) error {
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return nil
}

func (p *ApiVector) ToDatabase() (*Vector, error) {
	av := &Vector{}
	av.X = p.X
	av.Y = p.Y
	av.Angle = p.Angle
	av.Speed = p.Speed
	return av, nil
}

func (alog *ApiShuttleLog) FromDatabase(log *ShuttleLog) error {
	alog.VehicleID = log.VehicleID
	alog.Status = log.Status
	av := ApiVector{}
	av.FromDatabase(log.Location)
	alog.Location = av
	return nil
}

func handleLog(ctx *Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		switch r.Method {
		case "GET":
			id, err := getID(r, "id")
			if handleErr(w, err) {
				return
			}
			res, err := ctx.DB.SelectLatestLog(id)
			if handleErr(w, err) {
				return
			}
			ar := &ApiShuttleLog{}
			err = ar.FromDatabase(res)
			if handleErr(w, err) {
				return
			}
			err = sendResponse(w, ar)
			if handleErr(w, err) {
				return
			}
			measureTime(start, "Get Shuttle log")
		}
	}
}

func handleRoute(ctx *Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		switch r.Method {
		case "GET":
			id, err := getID(r, "name")
			if handleErr(w, err) {
				return
			}
			res, err := ctx.DB.SelectClosedRoute(id)
			if handleErr(w, err) {
				return
			}
			ar := &ApiClosedRoute{}
			err = ar.FromDatabase(res)
			if handleErr(w, err) {
				return
			}
			err = sendResponse(w, ar)
			if handleErr(w, err) {
				return
			}
			measureTime(start, "Get Route")
			break
		case "POST":
			decoder := json.NewDecoder(r.Body)
			route := &ApiClosedRoute{}
			err := decoder.Decode(route)
			if handleErr(w, err) {
				return
			}
			dbRoute, err := route.ToDatabase()
			if handleErr(w, err) {
				return
			}
			err = ctx.DB.InsertClosedRoute(dbRoute)
			if handleErr(w, err) {
				return
			}
			err = sendResponse(w, dbRoute)
			if handleErr(w, err) {
				return
			}
			measureTime(start, "POST Route")
			break
		default:
			handleErr(w, fmt.Errorf("%s Method not supported", r.Method))
		}
	}
}

func handleErrWithInfo(w http.ResponseWriter, err error, info string) bool {
	if err != nil {
		w.Write(Stat(ERROR, err.Error()+info))
		return true
	}
	return false
}

func handleErr(w http.ResponseWriter, err error) bool {
	return handleErrWithInfo(w, err, "")
}

func getID(r *http.Request, id string) (string, error) {
	q := r.URL.Query()
	str := q.Get(id)
	if str == "" {
		return "", errors.New("Invalid ID")
	}
	return str, nil
}

func validateToken(r *http.Request, token string) bool {
	// security feature to validate the token when posting
	return true
}

func sendResponse(w http.ResponseWriter, obj interface{}) error {
	res, err := json.Marshal(obj)
	if err != nil {
		return errors.New("Can't marshal")
	}
	w.Write(res)
	return nil
}

func Run(ctx *Context, config *Config) {
	fmt.Println("Running Shuttle server\n")
	// initialize router
	http.HandleFunc("/v1/shuttle", handleLog(ctx))
	http.HandleFunc("/v1/route", handleRoute(ctx))
	log.Fatal(http.ListenAndServe(config.LocalURL, nil))

	fmt.Println("End Shuttle server\n")
}
