#!/bin/bash

OUT_DIR=out
IMG_DIR=img
LOAD_START=25
LOAD_END=1000
LOAD_STEP=25
LOAD_STEP_DURATION=60s

N=1
NCORES=1.0
mkdir -p $OUT_DIR
mkdir -p $IMG_DIR

RPS_LEVELS=(1000)


for RPS in "${RPS_LEVELS[@]}"; do
  echo "start with RPS_LEVEL=$RPS"
  
    for ((RUN=1; RUN<=N; RUN++)); do
      OUT="$OUT_DIR/grpc_rps_${RPS}_n_${RUN}"
      ghz \
          --proto ../api/proto/embedder.proto \
          --call embedding.EmbeddingService.CreatePreferenceEmbedding \
          --insecure \
          --rps=$RPS \
          --connections=10    \
          --concurrency=10 \
          --duration=10s \
          --output=$OUT.txt \
          --format=summary \
          --data '{"text":"this is the test string"}' 0.0.0.0:50051

      python3 grpc_parse_ghz_summary.py $OUT.txt $OUT.json
  done


  python3 step_result_reducer.py grpc $RPS ./out/grpc_rps_${RPS}_reduced.json
  python3 plotter.py ./out/grpc_rps_${RPS}_reduced.json  ./img/grpc_rps_${RPS}_hist.png ${RPS}
done
