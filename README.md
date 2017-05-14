# YAST Server
##Yet another shuttle tracker
A golang API server targeting to replace current implementation of Shuttle tracker server.

##API List

| Type        | Request           | Response |
| ------------- |:-------------:| -----:|
| Shuttle | `GET /v1/shuttle?id=<shuttle id>` | latest shuttle location log |
| Route | `GET /v1/route?id=<route id>`      | an ordered list of map points on the map 