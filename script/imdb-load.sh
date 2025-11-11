#!/bin/bash
curl -L -o ./data/imdb-top-1000.zip\
  https://www.kaggle.com/api/v1/datasets/download/harshitshankhdhar/imdb-dataset-of-top-1000-MovieMetas-and-tv-shows

unzip ./data/imdb-top-1000.zip ./data