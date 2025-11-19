#!/bin/bash

OUT_DIR=increase_rps_out
IMG_DIR=img
RPS_VALUES=(500)
DURATION="60s"
CONTAINER_NAME=embedder-app
NCORES=1.0
N=1

mkdir -p $OUT_DIR
mkdir -p $OUT_DIR/grpc
mkdir -p $IMG_DIR

echo "Benchmark: RPS increase"
echo "RPS levels: ${RPS_VALUES[@]}"

echo "Begin gRPC test"
echo "Starting benchmark with $N runs per RPS level"
echo "RPS levels: ${RPS_VALUES[@]}"
for ((RUN=1; RUN<=N; RUN++)); do
    echo "run $RUN/$N"
    
    METRICS_FILE="$OUT_DIR/grpc/metrics_n_${RUN}_cores_${NCORES}.csv"
    ./metric_scraper.sh $CONTAINER_NAME $METRICS_FILE 1 & 
    MONITOR_PID=$!
    echo "scraper PID $MONITOR_PID"

    for RPS in "${RPS_VALUES[@]}"; do        
        OUTPUT_FILE="$OUT_DIR/grpc/rps_${RPS}_n_${RUN}_cores_${NCORES}.csv"
        ghz \
            --proto ../api/proto/embedder.proto \
            --call embedding.EmbeddingService.CreatePreferenceEmbedding \
            --insecure \
            --connections=10 \
            --concurrency=10 \
            --rps=$RPS \
            --duration=$DURATION \
            --output=$OUTPUT_FILE \
            --format=csv \
            --data '{"text":"this is the test string"}' 0.0.0.0:50051
    done
    
    kill $MONITOR_PID
    wait $MONITOR_PID 2>/dev/null
    pkill -f "metric_scraper.sh" 2>/dev/null
done
sed -i '/Unavailable/d' $OUT_DIR/grpc/*.csv
python3 average_latency.py increase_rps_out $N $NCORES grpc
python3 average_metrics.py increase_rps_out $N $NCORES grpc


