#!/bin/bash

python3 ./embedder/main.py
docker-compose -f docker-compose.prod.yml up --build