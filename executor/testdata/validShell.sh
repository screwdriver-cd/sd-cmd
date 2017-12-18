#!/bin/bash

echo "Hello World"

count=1

while [ $count -le $# ]
do
  eval echo '$'$count
  count=`expr $count + 1`
done
