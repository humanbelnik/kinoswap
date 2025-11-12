#!/bin/bash

OUT_DIR=out_grow
IMG_DIR=img
RPS_VALUES=($(seq 25 25 600))
DURATION="3s"

mkdir -p $OUT_DIR

echo "rps,duration_ms,status,error" > $OUT_DIR/grpc_combined.csv

for RPS in "${RPS_VALUES[@]}"; do
    echo "Testing RPS: $RPS"
    
    ghz \
        --proto ../api/proto/embedder.proto \
        --call embedding.EmbeddingService.CreatePreferenceEmbedding \
        --insecure \
        --connections=10 \
        --concurrency=10 \
        --rps=$RPS \
        --duration=$DURATION \
        --output=$OUT_DIR/grpc_rps_${RPS}.csv \
        --format=csv \
        --data '{"text":"this is the test string"}' 0.0.0.0:50051    
    echo "Completed RPS: $RPS"
done

for f in ./out_grow/grpc_*.csv; do head -n -10 "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"; done
python3 plotter_times.py img/grpc_timeline.png