package yast

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/keyboardnerd/yastserver/database"
	"github.com/keyboardnerd/yastserver/pkg"
)

var (
	logregex = regexp.MustCompile(
		`Vehicle ID:(\d+) lat:(-?\d+.\d+) lon:(-?\d+.\d+) dir:(\d+.\d+) spd:(\d+.\d+) lck:(\d+) time:(\d+) date:(\d+) trig:(\d+) eof`)
)

// Updater updates the database by an interval
type Updater struct {
	// fetch the data from remote API
	Fetcher Fetcher
	// database
	Database database.Database
	// intervae of fetching the data
	Interval int
}

// Fetcher fetches the data from remote API
type Fetcher struct {
	RemoteSite string
}

// RunUpdate fetches the data and insert into the database by based on some intervals
func (updater *Updater) RunUpdate() {
	fmt.Printf("run update... %#v\n", updater)
	updater.update(time.Now())
	for now := range time.Tick(time.Duration(updater.Interval) * time.Second) {
		updater.update(now)
	}
}

func (updater *Updater) update(now time.Time) {
	shuttleLog, err := updater.Fetcher.Pull()
	start := time.Now()
	if err != nil {
		// log error
		fmt.Printf("%v : %s", now, err.Error())
	} else {
		for _, log := range shuttleLog {
			func(x database.ShuttleLog) {
				err := updater.Database.InsertShuttleLog(&x)
				if err != nil {
					fmt.Printf("Unable to insert shuttle log to database %s\n", err.Error())
				}
			}(log)
		}
	}
	pkg.MeasureTime(start, fmt.Sprintf("database transaction, updated %d shuttles", len(shuttleLog)))
}

// Pull the data from upper stream, this is a blocking call
func (fetcher *Fetcher) Pull() ([]database.ShuttleLog, error) {
	// simple monitoring ( change to prometheus later )
	start := time.Now()
	// download the data from fetcher
	resp, err := http.Get(fetcher.RemoteSite)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	log, err := ParseShuttleLog(body)
	if err != nil {
		return nil, err
	}
	pkg.MeasureTime(start, "Pull remote shuttle log")
	return log, err
}

// ParseShuttleLog parses bytes received from the remote server and generate the database shutte log
func ParseShuttleLog(logslice []byte) ([]database.ShuttleLog, error) {
	var err error
	logVehicles := logregex.FindAllStringSubmatch(string(logslice), -1)
	if logVehicles == nil {
		err = fmt.Errorf("Failed to parse the response %s", logslice)
		return nil, err
	}
	rgshuttleLog := make([]database.ShuttleLog, len(logVehicles))
	for i, logVehicle := range logVehicles {
		// skip first one because it's the whole matched string
		log := database.ShuttleLog{}
		log.VehicleID = logVehicle[1]
		v := database.MapPoint{}
		v.X, err = strconv.ParseFloat(logVehicle[2], 10)
		if err != nil {
			return nil, err
		}
		v.Y, err = strconv.ParseFloat(logVehicle[3], 10)
		if err != nil {
			return nil, err
		}
		v.Angle, err = strconv.ParseFloat(logVehicle[4], 10)
		if err != nil {
			return nil, err
		}
		v.Speed, err = strconv.ParseFloat(logVehicle[5], 10)
		if err != nil {
			return nil, err
		}
		log.Location = &v
		log.Status = logVehicle[9]
		rgshuttleLog[i] = log
	}
	return rgshuttleLog, nil
}
