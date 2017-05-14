# YAST Server
## Yet another shuttle tracker
A golang API server targeting to replace current implementation of Shuttle tracker server.

## API Overview

| Type        | Request           | Response |
| ------------- |:-------------:| -----:|
| Shuttle | `GET /v1/shuttle?id=<shuttle id>` | latest shuttle location log |
| Route | `GET /v1/route?id=<route id>`      | an ordered list of map points on the map 
| Route | `POST /v1/route`      | post a new route to the database


## API Request/Response formats
notation: type & explanation

only `_stat` and `_info` will be returned on failure

~~~
Shuttle Get response on success; 
{
    "_stat": string & status of the response,
    "_info": string & additional information of the response, 
    "id" : string & external name of the vehicle,
    "location" : {
        "x" : float & longitude,
        "y" : float & latitude,
        "angle" : float & angle in degree,
        "speed" : float & speed in mph
    } & location of the shuttle in log,
    "stat" : string & status of the shuttle in log
}
~~~

~~~
Route Get/POST response
{
    "_stat": string & status of the response,
    "_info": string & additional information of the response, 
    "location" : [{
        "x" : float & longitude,
        "y" : float & latitude,
        "angle" : float & angle in degree,
        "speed" : float & speed in mph
    }] & ordered list of locations on the route,
    "name" : string & external name of the route
}
~~~

~~~
Route Post json
{
    "location" : [{
        "x" : float & longitude,
        "y" : float & latitude,
        "angle" : float & angle in degree,
        "speed" : float & speed in mph
    }] & ordered list of locations on the route,
    "name" : string & external name of the route ( should be unique )
}
~~~
