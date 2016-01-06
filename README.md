# TripPlanner

##Trip planner Using Uber Services using GO

The trip planner is a feature that will take a set of locations from the database and will then check against UBER’s price estimates API to suggest the best possible route in terms of costs and duration.

####API Used:
1.Google Map API
2.Uber Services

####Database Used:
 MongoDB

##Installation
#####To run the Trip Planner Service you need to install
```
1. mgo driver for MongoDb driver
  >  go get gopkg.in/mgo.v2
2. httprouter
   Download HttpRouter from GitHub and Run
  > go get
```
## Usage

Clone the repository CMPE-273-Assignment-3

###Start the  server:

```
cd CMPE-273-Assignment-3
go run TripPlanner.go
```
### 1. Plan a Trip

HTTP Request Used : POST

####Sample cURL command for POST Request:
```
> curl -H "Content-Type: application/json" -X POST -d '{"starting_from_location_id": "9879", "location_ids" : ["11066","48084","48088","46118"]}' http://localhost:8080/trips/
```

####Sample Response:
```
  {
"Id":9265,
"Status":"planning",
"Starting_from_location_id":"9879",
"Best_route_location_ids":["11066","48084","46118","48088"],
"Total_uber_costs":71,
"Total_uber_duration":3362,
"Total_distance":39.16
}
```

### 2. Get Existing Planned Trip

HTTP Request Used : GET

####Sample cURL command for GET Request:
```
  > curl -H "Content-Type: application/json" -X GET http://127.0.0.1:8080/trips/9265
```
####Sample Response:
```
 {
"Id":9265,
"Status":"planning",
"Starting_from_location_id":"9879",
"Best_route_location_ids":["11066","48084","46118","48088"],
"Total_uber_costs":71,
"Total_uber_duration":3362,
"Total_distance":39.16
}
```
### 3. Request Cab For Next Destination


HTTP Request Used : PUT

####Sample cURL command for PUT Request:
```
> curl -H "Content-Type: application/json" -X PUT http://localhost:8080/trips/2697/request
```
####Sample Response:
```
{
"Id":2697,
"Status":"requesting",
"Starting_from_location_id":"9879",
"Next_destination_location_id":"11066",
"Best_route_location_ids":["11066","48084","46118","48088"],
"Total_uber_costs":71,
"Total_uber_duration":3362,
"Total_distance":39.16,
"Uber_wait_time_eta":8
}
```
