package YAST

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Updater struct {
	Fetcher  Fetcher
	Database Database
	Interval int
}

func (updater *Updater) RunUpdate() {
	fmt.Printf("run update... %#v\n", updater)
	updater.update(time.Now())
	for now := range time.Tick(time.Duration(updater.Interval) * time.Second) {
		updater.update(now)
	}
}

func (updater *Updater) update(now time.Time) {
	fmt.Printf("Start Pulling Update at %v\n", time.Now())
	shuttleLog, err := updater.Fetcher.Pull()
	start := time.Now()
	if err != nil {
		// log error
		fmt.Printf("%v : %s", now, err.Error())
	} else {
		// insert into database in parallel
		for _, log := range shuttleLog {
			func(x ShuttleLog) {
				err := updater.Database.InsertShuttleLog(&x)
				if err != nil {
					fmt.Printf("Unable to insert shuttle log to database %v\n", x)
					panic(err.Error())
					return
				}
			}(log)
		}
	}
	fmt.Printf("Database Transaction costs: %v\n", time.Since(start))
}

type Fetcher struct {
	RemoteSite string
}

// Pull the data from upper stream, this is a blocking call
func (fetcher *Fetcher) Pull() ([]ShuttleLog, error) {
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
	// simple logging, use logrus later
	fmt.Printf("Pull Vehicle update costs: %v\n", time.Since(start))
	return log, err
}

var (
	logregex = regexp.MustCompile(
		`Vehicle ID:(\d+) lat:(-?\d+.\d+) lon:(-?\d+.\d+) dir:(\d+.\d+) spd:(\d+.\d+) lck:(\d+) time:(\d+) date:(\d+) trig:(\d+) eof`)
)

func ParseShuttleLog(logslice []byte) ([]ShuttleLog, error) {
	var err error
	logVehicles := logregex.FindAllStringSubmatch(string(logslice), -1)
	if logVehicles == nil {
		err = fmt.Errorf("Failed to parse the response %s", logslice)
		return nil, err
	}
	rgshuttleLog := make([]ShuttleLog, len(logVehicles))
	for i, logVehicle := range logVehicles {
		// skip first one because it's the whole matched string
		log := ShuttleLog{}
		log.VehicleID = logVehicle[1]
		v := Vector{}
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