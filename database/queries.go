package database

const (
	selectAllRouteName = `SELECT name FROM route`
	insertMapPoint     = `INSERT INTO map_point (longitude, latitude, angle, speed) VALUES ($1, $2, $3, $4) RETURNING id`
	// select or insert the shuttle's meta data if the shuttle is not found
	soiShuttleMeta = `WITH new_shuttle_meta AS ( 
							INSERT INTO shuttle_meta (remote_shuttle_id, shuttle_name) 
							SELECT CAST($1 AS VARCHAR), CAST($2 AS VARCHAR) 
							WHERE NOT EXISTS (SELECT remote_shuttle_id FROM shuttle_meta WHERE remote_shuttle_id=$1)
							RETURNING id)
						SELECT id FROM shuttle_meta WHERE remote_shuttle_id = $1
						UNION
						SELECT id FROM new_shuttle_meta`
	insertShuttleLog = `INSERT INTO shuttle_log (map_point_id, shuttle_meta_id, created_at) VALUES($1, $2, CURRENT_TIMESTAMP)`
	selectShuttleLog = ` 
					SELECT shuttle_log.id, shuttle_name, status, created_at, longitude, latitude, angle, speed
						FROM shuttle_log 
							LEFT JOIN shuttle_meta ON shuttle_log.shuttle_meta_id = shuttle_meta.id
							LEFT JOIN map_point ON shuttle_log.map_point_id = map_point.id
                     WHERE shuttle_meta.remote_shuttle_id = $1
						`
	insertRoutePath = `
		INSERT INTO route_path (route_id, map_point_id, ordering) VALUES ($1, $2, $3)
	`
	insertRouteInstance = `
		INSERT INTO route (name) VALUES ($1) RETURNING id
	`
	selectRoute = `
		SELECT route.id, longitude, latitude, angle, speed
		FROM map_point
		JOIN route_path ON route_path.map_point_id = map_point.id
		JOIN route ON route.id = route_path.route_id
		WHERE route.name = $1
		ORDER BY route_path.ordering
	`
	selectRouteMeta = `
		SELECT id FROM route WHERE name = $1
	`
)
