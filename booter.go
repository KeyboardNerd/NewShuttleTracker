package yast

import (
	"github.com/keyboardnerd/yastserver/api"
	"github.com/keyboardnerd/yastserver/database"
)

func Boot(config *api.Config) {
	// connect to database
	database := &database.PgSQL{URL: config.DbSrc}
	database.Open()
	defer database.Close()
	// initialize
	fetcher := Fetcher{RemoteSite: config.RemoteURL}
	updater := Updater{Fetcher: fetcher, Database: database, Interval: config.UpdaterInterval}
	// run updater async
	go updater.RunUpdate()
	// run api server
	ctx := &api.Context{database}
	api.Run(ctx, config)
}
