#!/bin/bash

CONTAINER_NAME=$1 
OUTPUT_FILE=$2     

echo "timestamp,cpu_percent,memory_usage,memory_limit,memory_percent" > $OUTPUT_FILE

cleanup() {
    echo "Metric scraper stopping..." >&2
    exit 0
}

trap cleanup SIGTERM SIGINT

get_timestamp() {
    date +%s%3N
}

while true; do
    stats=$(docker stats --no-stream --format "{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}}" $CONTAINER_NAME 2>/dev/null)
    if [ $? -eq 0 ] && [ ! -z "$stats" ]; then
        timestamp=$(get_timestamp)
        cpu_percent=$(echo $stats | cut -d',' -f1 | sed 's/%//')
        mem_usage=$(echo $stats | cut -d',' -f2 | awk '{print $1}')
        mem_limit=$(echo $stats | cut -d',' -f2 | awk '{print $3}')
        mem_percent=$(echo $stats | cut -d',' -f3 | sed 's/%//')
        
        echo "$timestamp,$cpu_percent,$mem_usage,$mem_limit,$mem_percent" >> $OUTPUT_FILE
    fi    
done

