package database

// TODO: use prometheus to track query time
const (

	// select or insert the shuttle's meta data if the shuttle is not found
	soiShuttle = `WITH new_shuttle AS ( 
							INSERT INTO shuttle (remote_name, name) 
							SELECT CAST($1 AS VARCHAR), CAST($2 AS VARCHAR)
							WHERE NOT EXISTS (SELECT remote_name FROM shuttle WHERE remote_name = $1)
							RETURNING id)
						SELECT id FROM shuttle WHERE remote_name = $1
						UNION
						SELECT id FROM new_shuttle`
	// insert
	insertRoutePath  = `INSERT INTO route_point (route_id, ordering, longitude, latitude, angle) VALUES ($1, $2, $3, $4, $5)`
	insertRoute      = `INSERT INTO route (name) VALUES ($1) RETURNING id`
	insertShuttleLog = `INSERT INTO shuttle_point (longitude, latitude, angle, speed, status, shuttle_id, created_at) 
							VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP) RETURNING id`
	// select
	selectAllRouteName   = `SELECT name FROM route`
	selectAllShuttleName = `SELECT name FROM shuttle`
	selectAllStopName    = `SELECT name FROM stop`

	selectShuttleLog = `SELECT name, status, created_at, longtitude, latitude, angle, speed
						FROM shuttle_point
							LEFT JOIN shuttle ON shuttle_point.shuttle_id = shuttle.id
						WHERE shuttle.name = $1`

	selectRoute = `SELECT route_point.id, route_id, longitude, latitude, angle, speed, ordering
							FROM route_point 
							JOIN route ON route.id = route_point.route_id
							WHERE route.name = $1
							ORDER BY route_point.ordering`
)
