#!/bin/bash
set -e
docker build -t superman_detector:latest .
image_id=`docker images | grep superman_detector | grep latest | awk '{print $3}'`
echo "ImageID: $image_id"
curr_dir=`pwd`

echo "Running the container"
docker run -v $curr_dir/databases:/go/databases -p 127.0.0.1:10000:10000 $image_id