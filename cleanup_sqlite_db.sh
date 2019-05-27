#!/bin/bash

curr_dir=`pwd`
echo "Changing to current directory: ${curr_dir}"
cd $curr_dir

./sqlite3 databases/login.db "delete from request"

