#!/bin/bash

CORES=$1

python3 plotter_latency.py \
    processed/grpc_latency_cores_$CORES.csv \
    processed/http_latency_cores_$CORES.csv \
    img/latency_cores_$CORES.png

python3 plotter_metrics.py \
    processed/grpc_metrics_cores_$CORES.csv \
    processed/http_metrics_cores_$CORES.csv \
    img/metrics_cores_$CORES.png  

# python3 plotter_latency.py processed/grpc_latency_cores_$NCORES.csv img/grpc_latency_core_$NCORES.png 
# python3 plotter_metrics.py processed/grpc_metrics_cores_$NCORES.csv img/grpc_metrics_cores_$NCORES.png