#!/bin/bash
file="bookings"
if [ -f "$file" ] ; then 
  echo "removing $file..."
  rm bookings
fi

echo "building $file..."
go build -o $file cmd/web/*.go

echo "launching $file..."
./$file
