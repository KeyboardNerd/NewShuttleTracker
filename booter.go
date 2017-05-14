package YAST

func Boot(config *Config) {
	// connect to database
	database := &PgSQL{Url: config.DbSrc}
	database.Open()
	defer database.Close()
	// initialize
	fetcher := Fetcher{RemoteSite: config.RemoteURL}
	updater := Updater{Fetcher: fetcher, Database: database, Interval: config.UpdaterInterval}
	// run updater async
	go updater.RunUpdate()
	// run api server
	ctx := &Context{database}
	Run(ctx, config)
}
