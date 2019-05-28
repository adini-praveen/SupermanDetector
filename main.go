package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/umahmood/haversine"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// NullString is an alias for sql.NullString data type
type NullString sql.NullString

// Scan implements the Scanner interface for NullString
func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

type Env struct {
	sqlDb *sql.DB
}

// json object to map the endpoint input data
type RequestInput struct {
	Event_UUID		string		`json:"event_uuid"`
	Username 		string		`json:"username"`
	Unix_timestamp	int64		`json:"unix_timestamp"`
	IP_Address  	string		`json:"ip_address"`
}

// geo location of the IP in the request
type currentGeo struct {
	Lat		float64	`json:"lat"`
	Lon		float64	`json:"lon"`
	Radius	uint16	`json:"radius"`
}

// json object representing the preceeding or succeeding IP
type ipResponse struct {
	Ip			string		`json:"ip,omitempty"'`
	Speed		*float32	`json:"speed,omitempty"`
	Lat			float64		`json:"lat,omitempty"`
	Lon			float64		`json:"lon,omitempty"`
	Radius		uint16		`json:"radius,omitempty"`
	Timestamp	int64		`json:"unix_timestamp,omitempty"`
}

// Error response json object
type errResponse struct {
	Error 	string 	`json:"error"`
}

// Response json object
type Response struct {
	CurrentGeo						currentGeo	`json:"currentGeo"`
	TravelToCurrentGeoSuspicious	*bool		`json:"travelToCurrentGeoSuspicious,omitempty"`
	TravelFromCurrentGeoSuspicious	*bool		`json:"travelFromCurrentGeoSuspicious,omitempty"`
	PrecedingIpAccess				ipResponse	`json:"precedingIpAccess,omitempty"`
	SubsequentIpAccess				ipResponse	`json:"subsequentIpAccess,omitempty"`
}

//var sqlDb *sql.DB
var tm *time.Time
var currHaversineCoord haversine.Coord

func (env *Env) home(writer http.ResponseWriter, req *http.Request){
	decoder := json.NewDecoder(req.Body)
	var input RequestInput
	writer.Header().Set("Content-Type", "application/json")
	err := decoder.Decode(&input)
	if err != nil {
		respondWithError(err.Error(), writer)
		fmt.Println("handling ", req.RequestURI, ": ", err)
		return
	}

	tm := time.Unix(input.Unix_timestamp, 0)

	fmt.Println("UUID: ", input.Event_UUID)
	fmt.Println("Username: ", input.Username)
	fmt.Println("IP Address: ", input.IP_Address)
	fmt.Println("Time: ", tm)

	// Check for valid IP address
	valid_ip := net.ParseIP(input.IP_Address)
	if valid_ip == nil {
		respondWithError("Invalid IP Address", writer)
		fmt.Println("Invalid IP Address: ", input.IP_Address)
		return
	}

	// building a temp table that appends row_number column which is used in join condition
	selectStatement := fmt.Sprintf(`with new_table
		AS (select uuid, username, ipaddress, date_time,ROW_NUMBER() OVER (order by date_time) row_no FROM request where username="%s")
		select t.uuid, t.username, t.ipaddress, t.date_time, t1.uuid, t1.username, t1.ipaddress, t1.date_time, t2.uuid, t2.username, t2.ipaddress, t2.date_time from (select * from new_table where date_time="%s" and ipaddress="%s") as t
			LEFT JOIN (select * from new_table) as t1 ON t1.row_no = t.row_no-1
				LEFT JOIN (select * from new_table) as t2 ON t2.row_no = t.row_no+1;`, input.Username, strconv.FormatInt(input.Unix_timestamp, 10), input.IP_Address)

	tx, err := env.sqlDb.Begin()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	// insert the request data into the database
	_, err = env.sqlDb.Exec("insert into request(uuid, username, ipaddress, date_time) values(?, ?, ?, ?)", input.Event_UUID, input.Username, input.IP_Address, input.Unix_timestamp)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	tx.Commit()

	// Defined the paramters as NullString to handle dereferencing issue with nil values returned from the database.
	type sqlrow struct {
		uuid 		NullString
		username	NullString
		ipaddress	NullString
		date_time	NullString
	}

	// variables to hold the current, preceeding and succeeding database records
	var t, t1, t2 sqlrow

	row := env.sqlDb.QueryRow(selectStatement)
	switch err := row.Scan(&t.uuid, &t.username, &t.ipaddress, &t.date_time, &t1.uuid, &t1.username, &t1.ipaddress, &t1.date_time, &t2.uuid, &t2.username, &t2.ipaddress, &t2.date_time); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		var current = currentGeo{} // json object referencing the request data geo location
		var preceeding = ipResponse{} // json object referencing the preceeding immediate request w.r.t the request data
		var succeeding = ipResponse{}// json object referencing the succeeding immediate request w.r.t the request data
		var resp = Response{} // response json that is returned to the end user

		var preceedingHaversineCoord haversine.Coord
		var succeedingHaversineCoord haversine.Coord

		current.Lat, current.Lon, current.Radius = GetLatitudeAndLongitude(input.IP_Address)
		if current.Lat != -10000 {
			resp.CurrentGeo = current
			currHaversineCoord = haversine.Coord{Lat:current.Lat, Lon:current.Lon}
		} else {
			respondWithError(fmt.Sprintf("Error retreiving geo location for ip %s", input.IP_Address), writer)
			return
		}

		// if t1.uuid is not null
		if t1.uuid.Valid {
			preceeding.Lat, preceeding.Lon, preceeding.Radius = GetLatitudeAndLongitude(t1.ipaddress.String)
			if preceeding.Lat != -10000 {
				// creating a bool and speed object so that the json keys can map to these objects so that they are not ignored
				// when displaying 0 or nil values because of omitempty flag set on the key
				tr := new(bool)
				speed := new(float32)
				tm1, _ := strconv.ParseInt(t1.date_time.String, 10, 64)
				preceeding.Timestamp = tm1
				preceeding.Ip = t1.ipaddress.String
				preceedingHaversineCoord = haversine.Coord{Lat: preceeding.Lat, Lon: preceeding.Lon}
				dist, _ := haversine.Distance(currHaversineCoord, preceedingHaversineCoord)

				// Deducting the accuracy radius for both the locations from the previous haversine distance, considering the
				// location can be anywhere within the radius. In this case i'm assuming its on the circle. Before deducting
				// convert radius in kilometers to miles by multiplying with conversion factor 0.6214
				finalDist := dist - (float64(preceeding.Radius) + float64(current.Radius)) * 0.6214
				if finalDist < 0 {
					finalDist = 0
					resp.TravelToCurrentGeoSuspicious = tr
				} else {

					tmDiff := tm.Sub(time.Unix(tm1, 0)).Hours()
					// if hour is less than 0 assume hour = 1
					if tmDiff == 0 {
						tmDiff = 1
					}

					// Setting the suspicion flag to true if the speed to travel from preceeding location to the current location is greater than 500
					*speed = float32(dist / tmDiff)

					if *speed > 500 {
						*tr = true
					}

					resp.TravelToCurrentGeoSuspicious = tr
				}
				preceeding.Speed = speed
				resp.PrecedingIpAccess = preceeding
			} else {
				respondWithError(fmt.Sprintf("Error retreiving geo location for preceeding ip %s", t1.ipaddress.String), writer)
				return
			}
		}

		// if t2.uuid is not null
		if t2.uuid.Valid {
			succeeding.Lat, succeeding.Lon, succeeding.Radius = GetLatitudeAndLongitude(t2.ipaddress.String)
			if succeeding.Lat != -10000 {
				tr := new(bool)
				speed := new(float32)
				tm1, _ := strconv.ParseInt(t2.date_time.String, 10, 64)
				succeeding.Timestamp = tm1
				succeeding.Ip = t2.ipaddress.String
				succeedingHaversineCoord = haversine.Coord{Lat: succeeding.Lat, Lon: succeeding.Lon}
				dist, _ := haversine.Distance(currHaversineCoord, succeedingHaversineCoord)
				// Deducting the accuracy radius for both the locations from the previous haversine distance, considering the
				// location can be anywhere within the radius. In this case i'm assuming its on the circle. Before deducting
				// convert radius in kilometers to miles by multiplying with conversion factor 0.6214
				finalDist := dist - (float64(succeeding.Radius) + float64(current.Radius)) * 0.6214
				if finalDist < 0 {
					finalDist = 0
					resp.TravelFromCurrentGeoSuspicious = tr
				} else {
					tmDiff := time.Unix(tm1, 0).Sub(tm).Hours()
					// if hour is less than 0 assume hour = 1
					if tmDiff == 0 {
						tmDiff = 1
					}

					// Setting the suspicion flag to true if the speed to travel from current location to subsequent location is greater than 500
					*speed = float32(finalDist / tmDiff)
					if *speed > 500 {
						*tr = true
					}
					resp.TravelFromCurrentGeoSuspicious = tr
				}
				succeeding.Speed = speed
				resp.SubsequentIpAccess = succeeding
			} else {
				respondWithError(fmt.Sprintf("Error retreiving geo location for subsequent ip %s", t1.ipaddress.String), writer)
				return
			}
		}

		respJson, respErr := json.Marshal(resp)

		if respErr != nil {
			log.Fatal(respErr)
			respondWithError(respErr.Error(), writer)
			return
		} else {
			// Setting header content-type to application/json
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			writer.Write(respJson)
			return
		}
	default:
		fmt.Println(err)
	}
}

// this function return latitude, longitude and radius give an IPAddress. If it couldnt find the ip it returns dummy value of -10000
func GetLatitudeAndLongitude(ip string) (lat, lon float64, radius uint16) {
	geoIpdb, err := geoip2.Open("databases/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer geoIpdb.Close()

	ipAddress := net.ParseIP(ip)
	record, err := geoIpdb.City(ipAddress)
	if err != nil {
		log.Fatal(err)
		return -10000, -10000, 65535
	}
	lat = record.Location.Latitude
	lon = record.Location.Longitude
	radius = record.Location.AccuracyRadius
	return
}

func respondWithError(str string, writer http.ResponseWriter) {
	errResp := errResponse{str}
	errRespJson, _ := json.Marshal(errResp)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(errRespJson)
}


func handleRequests() {
	// Opens a db connection to login.db that keeps track of all the incoming requests
	sqlDb, sqlErr := sql.Open("sqlite3", "databases/login.db")
	if sqlErr != nil {
		log.Fatal(sqlErr)
		panic(sqlErr)
	}
	env := &Env{sqlDb: sqlDb}
	defer env.sqlDb.Close()
	http.HandleFunc("/", env.home)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
	//defer sqlDb.Close()
	handleRequests()
}
