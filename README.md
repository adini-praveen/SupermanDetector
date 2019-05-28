# Superman Detector

This tool is used to track the user requests from across the globe and identifies if the account is compromised by calculating the speed it takes to travel from current location to the next location and also from current location to the previous location and comparing against the speed threshold value of 500 miles per hour.

### Assumptions
```
1. The suspicious flag is set to true, if the speed is greater than 500 miles per hour.
2. The accuracy radius for the both the locations is deducted from the haversine distance under 
   the assumption that the actual origin can be on the circle which reduces the distance by radius.
3. If distance after deducting radius is less than 0 we assume distance as 0.
4. If time to travel between geo locations is less than an hour we assume its 1 hour, since the 
   minimum distance will be the distance calculated in step 2.

```

### Prerequisites

This project uses the below software and golang modules.

```
sqlite3 - store the POST request data
github.com/oschwald/geoip2-golang - golang module that is used for IP lookup in the GeoLite2-City.mmdb database
github.com/umahmood/haversine - golang module that is used for calculating the haversine distance between 2 coordinates
github.com/mattn/go-sqlite3 - sqlite3 client to connect to sqlite database

```

### Compiling and Execution

The project is built in a golang docker container as a multistage build. The artifact from the first stage is copied into the next image and used as an entrypoint. The enclosed build.sh will compile and runs the docker container exposing the endpoint API on http://localhost:10000

###### Compilation and building docker image

The attached build script (build.sh) builds the docker image using the Dockerfile and runs the container. Below are the steps performed during the creation of the image.
```
1. The build initially pulls down the golang docker image and installs the required packages.
2. Copies the required golang modules from vendor folder into vendor folder on the docker container.
3. Builds the go module and creates a binary executable - SupermanDetector. By default go1.11.x and higher uses vendor modules.
4. Launches a new golang image and copies the binary from the previous step into the current one. 
5. Uses the binary as the Entrypoint and builds the final docker image

**Note** By default go 1.11.x and above uses the vendor folder for dependencies.
```

###### Execution 

```
The attached build script (build.sh) captures the imageID of the latest SupermanDetector 
and then runs the image by exposing the docker port 10000 and also mounts the 
database folder on the local machine as /go/databases.

```

## Running the tests

Execute runTests.sh script from the root directory.

##### Break down into tests

```
1. Test Invalid IP
2. IP lookup in geolite db should return valid coordinates for valid input values.
3. First POST request should return only current geo location and both preceeding & subsequent requests must be empty.
4. Two POST requests for the same user should return either preceeding or subsequent response based on the epoch timestamp along with current geo location.
5. On Three POST requests, the last POST request with epoch time between the other 2 requests should return both preceeding and subsequent response along with current geo location.
6. Running step 4 and 5, the response must include TravelToCurrentGeoSuspicious or TravelFromCurrentGeoSuspicious or both based on if preceeding or subsequent or both exists in the response.
```

## Miscellaneous

* **build.sh** - This script builds the docker image and runs a container using the image exposing the API endpoint at http://localhost:10000
* **runTests.sh** - This script executes the tests in main_tests.go file using the go test framework.
* **cleanup_sqlite_db.sh** - This deletes the entires in request table in databases/login.db. This db is used for storing incoming requests.
* **cleanup_docker_images.sh** - This deletes the orphaned docker images.
* **sqlite3** - The sqlite binary used for connecting to sqlite db. 

## Authors

* **Praveen Adini**

## References

* http://www.sqlitetutorial.net/sqlite-window-functions/sqlite-lag/
* https://github.com/umahmood/haversine
* https://github.com/oschwald/geoip2-golang
* https://docs.docker.com/develop/develop-images/multistage-build/
* https://blog.alexellis.io/golang-writing-unit-tests/

