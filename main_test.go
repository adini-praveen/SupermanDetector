package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var url = "http://localhost:10000"

func randomAlphanumeric(n int) string {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	//fmt.Println("called randomAlphanumeric")
	//fmt.Printf("Random String %s", string(b))
	return string(b)
}

func isEmpty(input ipResponse) bool {
	fmt.Printf("Checking for empty struct\n")
	flag := reflect.DeepEqual(input, ipResponse{})
	return flag
}

func TestInvalidIP(t *testing.T) {
	var data = new(RequestInput)
	var errResp = errResponse{}
	var eventId = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data.Event_UUID = eventId
	data.Username = "praveen"
	data.IP_Address = "9.11.255.22323"
	data.Unix_timestamp = 1514764800

	dataJson, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	response, err := http.Post(url, "application/josn", bytes.NewBuffer(dataJson))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	if err := json.NewDecoder(response.Body).Decode(&errResp); err != nil {
		t.Errorf("Error reading response body - %s\n", err)
	}

	if errResp.Error != "Invalid IP Address" {
		fmt.Println(errResp)
		t.Fail()
	}

}

func TestIPLookup(t *testing.T) {
	valid_ip1 := "5.38.127.221" // Dubai
	valid_ip2 := "41.31.255.255" // Bhutan
	lat, lon, rad := GetLatitudeAndLongitude(valid_ip1)
	if (lat == -10000 || lon == -10000 || rad == 65535) {
		t.Errorf("IP Address lookup failed")
	}

	lat, lon, rad = GetLatitudeAndLongitude(valid_ip2)
	if (lat == -10000 || lon == -10000 || rad == 65535) {
		t.Errorf("IP Address lookup failed")
	}
}

func TestFirstRequest(t *testing.T) {
	var data = new(RequestInput)
	var resp = new(Response)

	var eventId = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data.Event_UUID = eventId
	data.Username = randomAlphanumeric(10)
	data.IP_Address = "9.11.255.223"
	data.Unix_timestamp = 1514764800

	fmt.Printf("Event ID: %s", data.Event_UUID)
	fmt.Printf("Username: %s", data.Username)
	dataJson, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	response, err := http.Post(url, "application/josn", bytes.NewBuffer(dataJson))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		t.Errorf("Error reading response body - %s\n", err)
	}

	if resp.CurrentGeo.Lon == -10000 || resp.CurrentGeo.Lat == -10000 || resp.CurrentGeo.Radius == 65535 {
		t.Errorf("Error looking up geo location for IP - %s\n", data.IP_Address)
	}

	fmt.Printf("%+v\n",resp)
	fmt.Printf("Preceeding Request: %+v\n", resp.PrecedingIpAccess)
	fmt.Printf("Subsequent Request: %+v\n", resp.SubsequentIpAccess)
	if !isEmpty(resp.PrecedingIpAccess) || !isEmpty(resp.SubsequentIpAccess) ||
		resp.TravelToCurrentGeoSuspicious != nil || resp.TravelFromCurrentGeoSuspicious != nil {
		t.Errorf("The preceeding and subsequent requests should be empty for the first user POST request")
	}
}

func TestTwoRequest(t *testing.T) {
	var data = new(RequestInput)
	var data1 = new(RequestInput)
	var resp = new(Response)

	var eventId = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data.Event_UUID = eventId
	data.Username = randomAlphanumeric(10)
	data.IP_Address = "9.11.255.223"
	data.Unix_timestamp = 1514754800

	fmt.Printf("Event ID: %s\n", data.Event_UUID)
	fmt.Printf("Username: %s\n", data.Username)

	var eventId1 = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data1.Event_UUID = eventId1
	data1.Username = data.Username
	data1.IP_Address = "9.11.255.200"
	data1.Unix_timestamp = 1514764800
	fmt.Printf("Event ID1: %s\n", data1.Event_UUID)
	fmt.Printf("Username1: %s\n", data1.Username)

	dataJson, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	_, err = http.Post(url, "application/josn", bytes.NewBuffer(dataJson))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	dataJson1, err := json.Marshal(data1)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	response1, err := http.Post(url, "application/josn", bytes.NewBuffer(dataJson1))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	if err := json.NewDecoder(response1.Body).Decode(&resp); err != nil {
		t.Errorf("Error reading response body - %s\n", err)
	}

	if resp.CurrentGeo.Lon == -10000 || resp.CurrentGeo.Lat == -10000 || resp.CurrentGeo.Radius == 65535 {
		t.Errorf("Error looking up geo location for IP - %s\n", data.IP_Address)
	}

	fmt.Printf("Response object: %+v\n", resp)
	fmt.Printf("Preceeding Request: %+v\n", resp.PrecedingIpAccess)
	fmt.Printf("Subsequent Request: %+v\n", resp.SubsequentIpAccess)
	if isEmpty(resp.PrecedingIpAccess) ||
		!isEmpty(resp.SubsequentIpAccess) ||
		resp.TravelToCurrentGeoSuspicious == nil ||
		resp.TravelFromCurrentGeoSuspicious != nil  ||
		*(resp.TravelToCurrentGeoSuspicious) {
		t.Errorf("The preceeding response must not be empty and subsequent response must be empty")
	}
}

func TestThreeRequest(t *testing.T) {
	var data = new(RequestInput)
	var data1 = new(RequestInput)
	var data2 = new(RequestInput)
	var resp = new(Response)

	var eventId = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data.Event_UUID = eventId
	data.Username = randomAlphanumeric(10)
	data.IP_Address = "9.11.255.223" // USA
	data.Unix_timestamp = 1514730000

	fmt.Printf("Event ID: %s\n", data.Event_UUID)
	fmt.Printf("Username: %s\n", data.Username)

	dataJson, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	_, err = http.Post(url, "application/josn", bytes.NewBuffer(dataJson))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	var eventId1 = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data1.Event_UUID = eventId1
	data1.Username = data.Username
	data1.IP_Address = "9.11.255.200" // USA
	data1.Unix_timestamp = 1514764800
	fmt.Printf("Event ID1: %s\n", data1.Event_UUID)
	fmt.Printf("Username1: %s\n", data1.Username)

	dataJson1, err := json.Marshal(data1)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}
	_, err = http.Post(url, "application/josn", bytes.NewBuffer(dataJson1))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	var eventId2 = randomAlphanumeric(8) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(4) + "-" + randomAlphanumeric(12)
	data2.Event_UUID = eventId2
	data2.Username = data.Username
	data2.IP_Address = "41.72.223.250" // Bhutan
	data2.Unix_timestamp = 1514743800
	fmt.Printf("Event ID1: %s\n", data2.Event_UUID)
	fmt.Printf("Username1: %s\n", data2.Username)

	dataJson2, err := json.Marshal(data2)
	if err != nil {
		t.Errorf("Marshal Errror - %s\n", err)
		return
	}

	response, err := http.Post(url, "application/josn", bytes.NewBuffer(dataJson2))
	if err != nil {
		t.Errorf("The HTTP request failed with error %s\n", err)
	}

	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		t.Errorf("Error reading response body - %s\n", err)
	}

	if resp.CurrentGeo.Lon == -10000 || resp.CurrentGeo.Lat == -10000 || resp.CurrentGeo.Radius == 65535 {
		t.Errorf("Error looking up geo location for IP - %s\n", data.IP_Address)
	}

	fmt.Printf("Response object: %+v\n", resp)
	fmt.Printf("Preceeding Request: %+v\n", resp.PrecedingIpAccess)
	fmt.Printf("Subsequent Request: %+v\n", resp.SubsequentIpAccess)
	if *(resp.TravelFromCurrentGeoSuspicious) {
		fmt.Println("TravelFromCurrentGeoSuspicious: True")
	} else {
		fmt.Println("TravelFromCurrentGeoSuspicious: False")
	}
	if *(resp.TravelToCurrentGeoSuspicious) {
		fmt.Println("TravelToCurrentGeoSuspicious: True")
	} else {
		fmt.Println("TravelToCurrentGeoSuspicious: False")
	}
	if isEmpty(resp.PrecedingIpAccess) ||
		isEmpty(resp.SubsequentIpAccess) ||
		resp.TravelToCurrentGeoSuspicious == nil ||
		resp.TravelFromCurrentGeoSuspicious == nil ||
		!*(resp.TravelFromCurrentGeoSuspicious) ||
		!*(resp.TravelToCurrentGeoSuspicious) {
		t.Errorf("The preceeding and subsequent response cannot be empty")
	}
}


