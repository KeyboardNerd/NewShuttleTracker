# YAST Server
## Yet another shuttle tracker
A golang API server targeting to replace current implementation of Shuttle tracker server. The API is designed to be simple.

## Concept
Route is a referenced way of running between each stops. 
Shuttles are running on or not on the routes. 

## API ( to be implemented )
### Shuttle
#### GET /shuttle
Options

* active : return all active shuttle's latest raw information only

~~~
{
	"shuttles" : [
		{
			"name" : "deep blue",
			"location" : {
				"lat" : 123.3,
				"lon" : 234.4, 
			},
			"stop" : {
				"to" : {
					"name" : "east-colonie",
				},
				"from" : {
					"name" : "east-union",
				},
				"at" : {}
			},
			"heading" : 129,
			"status" : "active",
			"route" : "east",
		}
	]
}
~~~


### Route
#### GET /route
Options

* active : return all active routes only

~~~
{
	"routes" : [
		{
			"name" : "east",
			"location" : [{"lat": 1.0, "lon": 2.0 },...],
			"stops" : [
				{
					"name" : "east-colonie",
					"location" : {
						"lat" : 123.4,
						"lon" : 124.5
					},
					"order" : 0
				}
			]
			"status" : "active"
		},...
	]
}
~~~

#### POST /route
Add a route, including all stops and a few route points to be intrapolated

~~~
{
	"name" : "east",
	"location" : [{"lat": 1.0, "lon": 2.0},...],
	"closed route" : true,
	"interpolate" : true,
	"enabled time" : [
		{"start time" : "dec 2",
		 "end time" : "dec 3"
		 },...
	],
	"stops" : [
		{
			"name" : "east-colonie",
			"location" : {
				"lat" : 123.4,
				"lon" : 124.5
			},
			"snap to route" : true,
			"order": 0
		},...
	]
}
~~~

### ETA
#### GET /eta/shuttle/`:shuttle name`/stop/`:stop name`
Return ETA to that stop from this shuttle

~~~
{
	"name" : "deep blue",
	"stop" : {
		"to" : {
			"name" : "east-union",
			"eta" : "20 min"
		}
	},
	"route" : "east"
}
~~~