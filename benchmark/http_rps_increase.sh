#!/bin/bash

OUT_DIR=increase_rps_out
IMG_DIR=img
RPS_VALUES=($(seq 25 25 400))
DURATION="3s"
CONTAINER_NAME=embedder-app
NCORES=2.0
N=5

mkdir -p $OUT_DIR
mkdir -p $OUT_DIR/http
mkdir -p $IMG_DIR

echo "Benchmark: RPS increase (HTTP)"
echo "RPS levels: ${RPS_VALUES[@]}"

echo "Begin HTTP test"
echo "Starting benchmark with $N runs per RPS level"
echo "RPS levels: ${RPS_VALUES[@]}"

for ((RUN=1; RUN<=N; RUN++)); do
    echo "Run $RUN/$N"
    
    METRICS_FILE="$OUT_DIR/http/metrics_n_${RUN}_cores_${NCORES}.csv"
    ./metric_scraper.sh $CONTAINER_NAME $METRICS_FILE 1 & 
    MONITOR_PID=$!
    echo "Scraper PID: $MONITOR_PID"

    for RPS in "${RPS_VALUES[@]}"; do        
        OUTPUT_FILE="$OUT_DIR/http/rps_${RPS}_n_${RUN}_cores_${NCORES}.csv"
        echo "Testing RPS: $RPS"
        
        hey -z 2s -q $RPS -n 10 -c 10 -m POST -T "application/json" -d '{"text": "hello"}' -o csv http://localhost:5000/preference_embedding > $OUTPUT_FILE
            
        python3 hey_to_csv.py $OUTPUT_FILE $RPS
    done
    
    kill $MONITOR_PID
    wait $MONITOR_PID 2>/dev/null
    pkill -f "metric_scraper.sh" 2>/dev/null
    echo "Run $RUN completed"
done

python3 average_latency.py increase_rps_out $N $NCORES http
python3 average_metrics.py increase_rps_out $N $NCORES http

