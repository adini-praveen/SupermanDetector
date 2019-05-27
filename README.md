# Superman Detector

This tool is used to track the user requests from across the globe and identifies if the account is compromised by calculating the speed it takes to travel from current location to the next location and also from current location to the previous location and comparing against the speed threshold value of 500 miles per hour.

### Prerequisites

This project uses the below software and golang modules

```
sqlite3 - store the POST request data
github.com/oschwald/geoip2-golang - golang module that is used for IP lookup in the GeoLite2-City.mmdb database
github.com/umahmood/haversine - golang module that is used for calculating the haversine distance between 2 coordinates
github.com/mattn/go-sqlite3 - sqlite3 client to connect to sqlite database

```

### Compiling and Execution

The project is built in a golang docker container as a multistage build. The artifact from the first stage is copied into the next image and used as an entrypoint. The enclosed build.sh will compile and runs the docker container exposing the endpoint on http://localhost:10000

###### Compilation 

```
1. The build initially pulls down the golang docker image and installs the required packages.
2. It then copies the required golang modules from vendor folder into vendor folder on the docker container.
3. It then builds the go module and creates a binary executable - SupermanDetector
4. It then launches a new golang image and copies the binary from the previous step into the current one. 
5. It then launches the binary as the Entrypoint.
```

###### Execution 

```
The attached build script captures the imageID of the latest SupermanDetector and then runs the image by exposing the docker port 10000 and also mounts the database folder on the local machine as /go/databases 
```

## Running the tests

Execute runTests.sh script from the root directory.

##### Break down into tests

```
1. Test Invalid IP
2. IP lookup in geolite db should return valid coordinates for valid input values.
3. First POST request should return only current geo location and preceeding and subsequent requests should be empty.
4. Two POST requests for the same user should return either preceeding or subsequent response based on the epoch timestamp along with current geo location.
5. On Three POST requests, the last POST request with epoch time between the other 2 requests should return both preceeding and subsequent response along with current geo location.
6. Running step 4 and 5, the response must include TravelToCurrentGeoSuspicious or TravelFromCurrentGeoSuspicious or both based on if preceeding or subsequent or both exists in the response.
```

## Authors

* **Praveen Adini**

## References

* http://www.sqlitetutorial.net/sqlite-window-functions/sqlite-lag/
* https://github.com/umahmood/haversine
* https://github.com/oschwald/geoip2-golang
* https://docs.docker.com/develop/develop-images/multistage-build/
* https://blog.alexellis.io/golang-writing-unit-tests/