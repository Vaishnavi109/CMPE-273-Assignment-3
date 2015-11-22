package main
import (
	
	"encoding/json"
    "fmt"
    "net/http"
    "github.com/julienschmidt/httprouter"	
    "io/ioutil"
   // s "strings"
    "math/rand"
    "strconv"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    "time"
    "bytes"
)
type ArgsForRequest struct {
      Starting_from_location_id string  
      Location_ids []string
}
type ErrorHandling struct{
	Status string
}
type Response struct{
	Id int64
	Status string
	Starting_from_location_id string
	Best_route_location_ids []string
	Total_uber_costs float64
	Total_uber_duration float64
	Total_distance float64
}
type PutRequestResponse struct{
     Id int64
     Status string
     Starting_from_location_id string
     Next_destination_location_id string
     Best_route_location_ids [] string
     Total_uber_costs float64
     Total_uber_duration float64
  	 Total_distance float64
     Uber_wait_time_eta float64
}

type PutRequestToDb struct{
     Id int64
     Status string
     Starting_from_location_id string
     Next_destination_location_id string
     Best_route_location_ids [] string
     Total_uber_costs float64
     Total_uber_duration float64
  	 Total_distance float64
     Uber_wait_time_eta float64
     Counter int
}
type Data struct{
	
	Id string `json:"id"`
	Name string `json:"name"`
    Address string `json:"address"`
    City string `json:"city"`
    State string `json:"state"`
    Zip string `json:"zip"`
    Coordinate Coor`json:"coordinate"`
  }
type Coor struct {
        Lat string `json:"lat"`
        Lng string `json:"lng"`
    }

var total_estimate float64
var total_duration float64
var total_distance float64

func PlanTrips(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var request ArgsForRequest
	response := Response{}
	
	json.NewDecoder(req.Body).Decode(&request)
	fmt.Println("IN PlanTrips")
	
	NewNearestLocation := getShortestRoute(request.Starting_from_location_id,request.Location_ids)
	response.Best_route_location_ids= append(response.Best_route_location_ids,NewNearestLocation)
	NewArray := RemoveFromArray(NewNearestLocation,request.Location_ids)
	
	for j:=0;len(response.Best_route_location_ids)!= len(request.Location_ids);j++{
		NewNearestLocation = getShortestRoute(NewNearestLocation,NewArray)
		response.Best_route_location_ids= append(response.Best_route_location_ids,NewNearestLocation)
		NewArray = RemoveFromArray(NewNearestLocation,NewArray)
		
	}
	fmt.Println(len(response.Best_route_location_ids))
	ReturnToStart(response.Best_route_location_ids[len(response.Best_route_location_ids)-1],request.Starting_from_location_id)
	fmt.Println(total_distance)
	fmt.Println(total_duration)
	fmt.Println(total_estimate)
	rand.Seed(time.Now().UTC().UnixNano())
	response.Id = rand.Int63n(9999)
	response.Status = "planning"
	response.Starting_from_location_id = request.Starting_from_location_id
	response.Total_uber_costs = total_estimate
	response.Total_uber_duration = total_duration
	response.Total_distance = total_distance
	
	uj, _ := json.Marshal(response)
    fmt.Fprintf(w, "%s",uj)

    mongoDBDialInfo := &mgo.DialInfo{
	Addrs:    []string{"ds043694.mongolab.com:43694"},
	Timeout:  60 * time.Second,
	Database: "locations",
	Username: "Admin",
	Password: "Admin123",
}    

    session, err := mgo.DialWithInfo(mongoDBDialInfo)
        if err != nil {
                panic(err)
        }
        defer session.Close()

        session.SetMode(mgo.Monotonic, true)
        c := session.DB("locations").C("TripDetails")
        err = c.Insert(response)
        if err != nil {
                log.Fatal(err)
        }
        
        total_distance = 0.00
        total_duration = 0.00
        total_estimate = 0.00
    	

}
func ReturnToStart(StartLocation string,EndLocation string){
	    mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{"ds043694.mongolab.com:43694"},
		Timeout:  60 * time.Second,
		Database: "locations",
		Username: "Admin",
		Password: "Admin123",
	    }    

	    session, err := mgo.DialWithInfo(mongoDBDialInfo)
	    if err != nil {
	        panic(err)
	    }
	    defer session.Close()

	    session.SetMode(mgo.Monotonic, true)
	    c := session.DB("locations").C("LocDetails")
	    startresult := Data{}
	    endresult := Data{}
		err = c.Find(bson.M{"id": StartLocation}).One(&startresult)
		if err != nil {
			panic(err)}

		err = c.Find(bson.M{"id": EndLocation}).One(&endresult)
		if err != nil {
			panic(err)}
		start_latitude := startresult.Coordinate.Lat
		start_longitude :=startresult.Coordinate.Lng	
		resp, err := http.Get("https://sandbox-api.uber.com/v1/estimates/price?start_latitude="+start_latitude+"&start_longitude="+start_longitude+"&end_latitude="+endresult.Coordinate.Lat+"&end_longitude="+endresult.Coordinate.Lng+"&server_token=o0fHM9H2CMsFq2wOBD2gYuoAe1V0MTjs5pYYY191")
		if err != nil {
					// handle error
					fmt.Println("Error")
				}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
	
    	var f interface{}
  		json.Unmarshal(body,&f)
    	//fmt.Println(error)
    	mRes := f.(map[string]interface{})["prices"]

    	mRes0 := mRes.([]interface{})[0]
    	//fmt.Println(mRes0)

    	estimate := mRes0.(map[string]interface{})["low_estimate"].(float64)
    	duration := mRes0.(map[string]interface{})["duration"].(float64)
    	distance := mRes0.(map[string]interface{})["distance"].(float64)

   		total_estimate+=estimate
	   	total_duration+=duration
   		total_distance+=distance


}
func RemoveFromArray(Location string , Location_ids[] string) []string{
	for i:=0;i<len(Location_ids);i++{
		if Location_ids[i]==Location{
			Location_ids = append(Location_ids[:i],Location_ids[i+1:]...)
			break
		}
			

	}
	return Location_ids

}
func getShortestRoute(startLocation string , Location_ids[] string) string{
	result := Data{}
	output := Data{}
	var LocationId string
	var estimate float64
	var duration float64
	var distance float64
	IdToGet := startLocation

	mongoDBDialInfo := &mgo.DialInfo{
	Addrs:    []string{"ds043694.mongolab.com:43694"},
	Timeout:  60 * time.Second,
	Database: "locations",
	Username: "Admin",
	Password: "Admin123",
    }    

    session, err := mgo.DialWithInfo(mongoDBDialInfo)
    if err != nil {
        panic(err)
    }
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)
    c := session.DB("locations").C("LocDetails")

	err = c.Find(bson.M{"id": IdToGet}).One(&result)
	if err != nil {
		panic(err)
	}
	start_latitude := result.Coordinate.Lat
	start_longitude :=result.Coordinate.Lng
	nearestPlaceEstimateFromStart := 9999.00
	nearestPlaceDurationFromStart := 9999.00
	nearestPlaceDistanceFromStart := 9999.00
	for i:=0;i<len(Location_ids);i++{
		err = c.Find(bson.M{"id": Location_ids[i]}).One(&output)
		if err != nil {
			panic(err)
		}
	
		resp, err := http.Get("https://sandbox-api.uber.com/v1/estimates/price?start_latitude="+start_latitude+"&start_longitude="+start_longitude+"&end_latitude="+output.Coordinate.Lat+"&end_longitude="+output.Coordinate.Lng+"&server_token=o0fHM9H2CMsFq2wOBD2gYuoAe1V0MTjs5pYYY191")
		if err != nil {
					// handle error
					fmt.Println("Error")
				}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
	
    	var f interface{}
  		json.Unmarshal(body,&f)
    	//fmt.Println(error)
    	mRes := f.(map[string]interface{})["prices"]

    	mRes0 := mRes.([]interface{})[0]
    	//fmt.Println(f)

    	estimate = mRes0.(map[string]interface{})["low_estimate"].(float64)
    	duration = mRes0.(map[string]interface{})["duration"].(float64)
    	distance = mRes0.(map[string]interface{})["distance"].(float64)
    	//fmt.Println("Route",startLocation,Location_ids[i])
    	//fmt.Println(estimate)
   		//fmt.Println("Low Estimate")
   		if estimate < nearestPlaceEstimateFromStart{
   			LocationId = Location_ids[i]
   			nearestPlaceEstimateFromStart = estimate
   			nearestPlaceDurationFromStart = duration
   			nearestPlaceDistanceFromStart = distance
   		} 
   		if estimate == nearestPlaceEstimateFromStart{
   			if duration < nearestPlaceDurationFromStart{
   				LocationId = Location_ids[i]
	   			nearestPlaceDurationFromStart = duration
	   			nearestPlaceEstimateFromStart = estimate
				nearestPlaceDistanceFromStart = distance
   			}
   			   		}
   	
	}
	fmt.Println(estimate)
	fmt.Println(duration)
	total_estimate+=nearestPlaceEstimateFromStart
   	total_duration+=nearestPlaceDurationFromStart
   	total_distance+=nearestPlaceDistanceFromStart
   	
	return LocationId
}
func GetTripDetails(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	IdToGet := p.ByName("trip_id")
	result := Response{}
	mongoDBDialInfo := &mgo.DialInfo{
	Addrs:    []string{"ds043694.mongolab.com:43694"},
	Timeout:  60 * time.Second,
	Database: "locations",
	Username: "Admin",
	Password: "Admin123",
}    

    session, err := mgo.DialWithInfo(mongoDBDialInfo)
    if err != nil {
            panic(err)
    }
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)
    c := session.DB("locations").C("TripDetails")
    val, err := strconv.Atoi(IdToGet)
	err = c.Find(bson.M{"id": val}).One(&result)
	if err != nil {
		panic(err)
	}
	var response Response
	
	response.Id = int64(val)
	response.Status = result.Status
	response.Starting_from_location_id = result.Starting_from_location_id
	response.Best_route_location_ids = result.Best_route_location_ids
	response.Total_uber_costs = result.Total_uber_costs
	response.Total_uber_duration = result.Total_uber_duration
	response.Total_distance = result.Total_distance
	uj, _ := json.Marshal(response)
    fmt.Fprintf(w, "%s",uj)

}
func RequestUber(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	//product_id := "04a497f5-380d-47f2-bf1b-ad4cfdcb51f2"
	accessToken :="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicmVxdWVzdCJdLCJzdWIiOiJmNzJkM2IzOS1jOTJkLTQxODgtOTQ2YS0zOWIyZGFjYjdhNTMiLCJpc3MiOiJ1YmVyLXVzMSIsImp0aSI6IjM0MjhiNDJkLTU4YzEtNDczMC1iNTI2LWMxM2Y0YjU3NDQwNCIsImV4cCI6MTQ1MDUxOTE3NSwiaWF0IjoxNDQ3OTI3MTc1LCJ1YWN0IjoiMjJUWVpXRkw1ZW1QRm1abjlhTTlxNUdXbkw4QXFKIiwibmJmIjoxNDQ3OTI3MDg1LCJhdWQiOiJMRUNuc1ozSV9YOW12Tmh6TDYzdUZPNmpyYkt4b3pscCJ9.Ksd8HuzGalGIhZqMDzYqjtQLT_oGlgrTVsW7DUiAHpSjqKzbfU3bQY_0rOM6mJaTwWrozVmADd5TZy8Culv0swU5Vpxz6muaLYv7z-D-vmPY66acX1FXDmB7QfyWniTe69BkSHANkz7XV0roJ4OYVjazC2AGP8k1NmkzPr3JBaSc1nvnXkBKpxmEMDQZ5l9cnKvN-L8DtD9CUPKTMKc2dcsq90O896-py9AHpY3YhGPguZFkn6Tkcrg5-XIqAuK6ksxXvkwtpBaYhaFBiNZG9xphIhqH_XOffdrATjrxqfwrT_bKMamLQZVammf7PwU8rYub8u2H83GeYAZblWHPiA"
	IdToGet := p.ByName("trip_id")
	result := Response{}
	startloc := Data{}
	dest := Data{}
	ResponseForDb := PutRequestToDb{}
	Response := PutRequestResponse{}
	var counter int
	var eta float64
	ExistingData := PutRequestToDb{}
	var Route []string	
	var flag int
	mongoDBDialInfo := &mgo.DialInfo{
	Addrs:    []string{"ds043694.mongolab.com:43694"},
	Timeout:  60 * time.Second,
	Database: "locations",
	Username: "Admin",
	Password: "Admin123",
	}    

    session, err := mgo.DialWithInfo(mongoDBDialInfo)
    if err != nil {
            panic(err)
    }
    defer session.Close()
    fmt.Println("In Put")
    session.SetMode(mgo.Monotonic, true)
    c := session.DB("locations").C("TripDetails")
    c2 := session.DB("locations").C("UpdatedDestinations")
    val, err := strconv.Atoi(IdToGet)
    err = c2.Find(bson.M{"id": val}).One(&ExistingData)
    if(err==nil){
    	counter = ExistingData.Counter
    	
    	flag = 1
    	
    }else{
    	counter = 0
    	flag = 0
    	
    }



    err = c.Find(bson.M{"id": val}).One(&result)
	if err != nil {
		panic(err)
	}
	starting_location := result.Starting_from_location_id
	Route=append(Route,starting_location)
	for i:=0;i<len(result.Best_route_location_ids);i++{
		Route=append(Route,result.Best_route_location_ids[i])
	}
	
		
		 c1 := session.DB("locations").C("LocDetails")
		 err = c1.Find(bson.M{"id": Route[counter]}).One(&startloc)
		 fmt.Println("In Put5")
		 if(counter>=len(Route)-1){
		 	fmt.Println("Last Location Reached")
		 	errorhandling := ErrorHandling{}
		 	counter =0
		 	errorhandling.Status="Last Destination- Trip Over" 
		 	uj, _ := json.Marshal(errorhandling)
		    fmt.Fprintf(w, "%s",uj)
		    ResponseForDb.Id = int64(val)
			ResponseForDb.Status = "requesting"
			ResponseForDb.Starting_from_location_id = starting_location
			ResponseForDb.Next_destination_location_id = Route[counter+1]
			ResponseForDb.Best_route_location_ids = result.Best_route_location_ids
			ResponseForDb.Total_uber_costs = result.Total_uber_costs
			ResponseForDb.Total_uber_duration = result.Total_uber_duration
			ResponseForDb.Total_distance = result.Total_distance
			ResponseForDb.Uber_wait_time_eta = eta
			ResponseForDb.Counter = counter	
			err = c2.Update(ExistingData, ResponseForDb)
			if err!=nil{
				fmt.Println("Error")
				}
		 	
		 }else{
		 	
			 err = c1.Find(bson.M{"id": Route[counter+1]}).One(&dest)
			 start_latitude := startloc.Coordinate.Lat
			 start_longitude :=startloc.Coordinate.Lng
			 end_latitude := dest.Coordinate.Lat
			 end_longitude := dest.Coordinate.Lng

			 Productresp, err := http.Get("https://sandbox-api.uber.com/v1/products?latitude="+start_latitude+"&longitude="+start_longitude+"&access_token="+accessToken)
			 
			 if err != nil {
					// handle error
					fmt.Println("Error")
				}
			 defer Productresp.Body.Close()
			 body, _ := ioutil.ReadAll(Productresp.Body)
	
    		 var prod interface{}
  			 json.Unmarshal(body,&prod)
  			 //fmt.Println(prod)
  			 mRes := prod.(map[string]interface{})["products"]

    		 mRes0 := mRes.([]interface{})[0]
    	
    	     product_id := mRes0.(map[string]interface{})["product_id"].(string)
    	     


			 fmt.Println(start_latitude,start_longitude,end_latitude,end_longitude)
			 urlParams,_ := json.Marshal(map[string] interface{}{
			 	"product_id":product_id,"start_latitude":start_latitude,"start_longitude":start_longitude,"end_latitude":end_latitude,"end_longitude":end_longitude})
			 request,err :=http.NewRequest("POST","https://sandbox-api.uber.com/v1/requests",bytes.NewBuffer(urlParams))
			 request.Header.Set("Content-Type","application/json")
			 request.Header.Set("Authorization","Bearer "+accessToken)
		 	 HttpClient := &http.Client{}
		 	 UberResponse,err := HttpClient.Do(request)
		 	 if err!=nil{
		 	 	fmt.Println("Error",err)
		 	 }else{
		 	 	defer UberResponse.Body.Close()
		 	 	var f interface{}
		 	 	definition,err := ioutil.ReadAll(UberResponse.Body)
		 	 	if(err!=nil){
		 	 		fmt.Println(err)
		 	 	}else{
		 	 		json.Unmarshal(definition,&f)
		 	 		
		 	 		request_id := f.(map[string]interface{})["request_id"].(string)
		 	 		eta = f.(map[string]interface{})["eta"].(float64)
		 	 		

		 	 		putParams,_:=json.Marshal(map[string] interface{}{"status":"completed"})
		 	 		url:="https://sandbox-api.uber.com/v1/sandbox/requests/"+request_id
		 	 		
		 	 		req,err := http.NewRequest("PUT",url,bytes.NewBuffer(putParams))
		 	 		req.Header.Set("Content-Type","application/json")
			 		req.Header.Set("Authorization","Bearer "+accessToken)
			 		Client := &http.Client{}
		 	 		resp,err:=Client.Do(req)
		 	 		if(err!=nil){
		 	 			fmt.Println("Error",err)
		 	 		}else{
		 	 			
		 	 			fmt.Println(resp.StatusCode)
		 	 		}
		 	 		

					Response.Id = int64(val)
					Response.Status = "requesting"
					Response.Starting_from_location_id = starting_location
					fmt.Println(counter)
					Response.Next_destination_location_id = Route[counter+1]
					Response.Best_route_location_ids = result.Best_route_location_ids
					Response.Total_uber_costs = result.Total_uber_costs
					Response.Total_uber_duration = result.Total_uber_duration
					Response.Total_distance = result.Total_distance
					Response.Uber_wait_time_eta = eta
					uj, _ := json.Marshal(Response)

				    fmt.Fprintf(w, "%s",uj)
				    
				    if flag == 0{
				    	ResponseForDb.Id = int64(val)
						ResponseForDb.Status = "requesting"
						ResponseForDb.Starting_from_location_id = starting_location
						ResponseForDb.Next_destination_location_id = Route[counter+1]
						ResponseForDb.Best_route_location_ids = result.Best_route_location_ids
						ResponseForDb.Total_uber_costs = result.Total_uber_costs
						ResponseForDb.Total_uber_duration = result.Total_uber_duration
						ResponseForDb.Total_distance = result.Total_distance
						ResponseForDb.Uber_wait_time_eta = eta
						counter++
						ResponseForDb.Counter = counter	
						err = c2.Insert(ResponseForDb)

						if err!=nil{
							fmt.Println("Error")
						}
				    }else{
				    	
				    	ResponseForDb.Id = int64(val)
						ResponseForDb.Status = "requesting"
						ResponseForDb.Starting_from_location_id = starting_location
						ResponseForDb.Next_destination_location_id = Route[counter+1]
						ResponseForDb.Best_route_location_ids = result.Best_route_location_ids
						ResponseForDb.Total_uber_costs = result.Total_uber_costs
						ResponseForDb.Total_uber_duration = result.Total_uber_duration
						ResponseForDb.Total_distance = result.Total_distance
						ResponseForDb.Uber_wait_time_eta = eta
						counter++
						ResponseForDb.Counter = counter	
						err = c2.Update(ExistingData, ResponseForDb)
						if err!=nil{
							fmt.Println("Error")
						}
				    }
		 	 		
		 	 	}
		 	 
		 }

		
	}



}
func main() {
    mux := httprouter.New()
    mux.POST("/trips", PlanTrips)

    mux.GET("/trips/:trip_id",GetTripDetails)
    mux.PUT("/trips/:trip_id/request",RequestUber)
    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }
    server.ListenAndServe()
}
