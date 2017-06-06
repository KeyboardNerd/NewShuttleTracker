package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/keyboardnerd/yastserver/pkg"
)

const (
	ERROR = "error"
	OK    = "ok"
)

func handleList(ctx *Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		switch r.Method {
		case "GET":
			q := r.URL.Query()
			switch q.Get("category") {
			case "shuttle":
				break
			case "stop":
				break
			case "route":
				break
			default:
				handleErr(w, fmt.Errorf("invalid category"))
			}
		}
	}
}

func handleHTTP(route string, f handleFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode, err := f(w, r)
		pkg.MeasureTime(start, route)
		if handleErr(w, err) {
			return
		}
	}
}
func (s *srvContext) getLatestShuttleLog(w http.ResponseWriter, r *http.Request) (int, error) {
	start := time.Now()
	id, err := getID(r, "id")
	if err != nil {
		http.Status
	}
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
			res, err := ctx.DB.SelectRoute(id)
			if handleErr(w, err) {
				return
			}
			ar := &ApiRoute{}
			err = ar.FromDatabase(res)
			if handleErr(w, err) {
				return
			}
			err = sendResponse(w, ar)
			if handleErr(w, err) {
				return
			}
			pkg.MeasureTime(start, "Get Route")
			break
		case "POST":
			decoder := json.NewDecoder(r.Body)
			route := &ApiRoute{}
			err := decoder.Decode(route)
			if handleErr(w, err) {
				return
			}
			dbRoute, err := route.ToDatabase()
			if handleErr(w, err) {
				return
			}
			err = ctx.DB.InsertRoute(dbRoute)
			if handleErr(w, err) {
				return
			}
			err = sendResponse(w, dbRoute)
			if handleErr(w, err) {
				return
			}
			pkg.MeasureTime(start, "POST Route")
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
	log.Println("Running Shuttle server")
	r := mux.NewRouter()
	r.HandleFunc("/shuttle/{id}").Methods("GET")
	// initialize router
	http.HandleFunc("/v1/shuttle", handleLog(ctx))
	http.HandleFunc("/v1/list", handleList(ctx))
	http.HandleFunc("/v1/route", handleRoute(ctx))
	log.Fatal(http.ListenAndServe(config.LocalURL, nil))

	fmt.Println("End Shuttle server\n")
}
