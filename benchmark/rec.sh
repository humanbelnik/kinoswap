#!/bin/bash

OUT_DIR=recovery_out
IMG_DIR=img
WARMUP_RPS=1200
WARMUP_DURATION="15s"
TARGET_RPS=560
TARGET_LATENCY=500
CONTAINER_NAME=ppo-api
NCORES=1.0
N=1

rm -rf $OUT_DIR
mkdir -p $OUT_DIR
mkdir -p $OUT_DIR/http
mkdir -p $IMG_DIR

echo "Benchmark: Recovery Time Test"
echo "Warmup: $WARMUP_RPS RPS for $WARMUP_DURATION"
echo "Target: $TARGET_RPS RPS until latency <= ${TARGET_LATENCY}ms"

echo "Begin recovery test"
echo "Starting benchmark with $N runs"

for ((RUN=1; RUN<=N; RUN++)); do
    echo "Run $RUN/$N"
    
    METRICS_FILE="$OUT_DIR/http/metrics_n_${RUN}_cores_${NCORES}.csv"
    ./metric_scraper.sh $CONTAINER_NAME $METRICS_FILE 1 & 
    MONITOR_PID=$!

    RECOVERY_FILE="$OUT_DIR/http/recovery_time_n_${RUN}_cores_${NCORES}.csv"
    echo "rps,recovery_time_ms" > $RECOVERY_FILE
    WARMUP_OUTPUT="$OUT_DIR/http/warmup_rps_${WARMUP_RPS}_n_${RUN}_cores_${NCORES}.csv"
    
    CONCURRENT_WARMUP=$(( WARMUP_RPS / 50 ))
    if [ $CONCURRENT_WARMUP -lt 10 ]; then
        CONCURRENT_WARMUP=10
    fi
    
    CONCURRENT_TARGET=$(( TARGET_RPS / 50 ))
    if [ $CONCURRENT_TARGET -lt 10 ]; then
        CONCURRENT_TARGET=10
    fi

    RECOVERY_OUTPUT="$OUT_DIR/http/recovery_rps_${TARGET_RPS}_n_${RUN}_cores_${NCORES}.csv"

    hey -z $WARMUP_DURATION -q $WARMUP_RPS -c $CONCURRENT_WARMUP -m POST \
        -T "application/json" \
        -m POST -T "application/json" -d '{"text": "hello"}' -o csv http://localhost:5000/preference_embedding > $WARMUP_OUTPUT
    START_TIME=$(date +%s%3N)
    RECOVERY_TIME=0
    sleep 0.01
    hey -z 100s -q $TARGET_RPS -c $CONCURRENT_TARGET -m POST \
        -T "application/json" \
        -m POST -T "application/json" -d '{"text": "hello"}' -o csv http://localhost:5000/preference_embedding  > $RECOVERY_OUTPUT &
    HEY_PID=$!

    while ps -p $HEY_PID > /dev/null; do
        CURRENT_TIME=$(date +%s%3N)
        ELAPSED=$((CURRENT_TIME - START_TIME))
        echo $ELAPSED
        
        kill $HEY_PID 2>/dev/null
        wait $HEY_PID 2>/dev/null
        RECOVERY_TIME=$ELAPSED
        echo "Recovery achieved at ${RECOVERY_TIME}ms"
        echo "$TARGET_RPS,$RECOVERY_TIME" >> $RECOVERY_FILE
        break
        
        sleep 1
    done

    python3 hey_to_csv.py $WARMUP_OUTPUT $WARMUP_RPS
    python3 hey_to_csv.py $RECOVERY_OUTPUT $TARGET_RPS

    if [ $RECOVERY_TIME -eq 0 ]; then
        RECOVERY_TIME=60000
        echo "$TARGET_RPS,$RECOVERY_TIME" >> $RECOVERY_FILE
        echo "Recovery not achieved within 60 seconds"
    fi
    
    kill $MONITOR_PID 2>/dev/null
    wait $MONITOR_PID 2>/dev/null
    pkill -f "metric_scraper.sh" 2>/dev/null
    echo "Run $RUN completed. Recovery time: ${RECOVERY_TIME}ms"

done
