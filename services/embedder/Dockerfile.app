FROM python:3.11-slim as builder

RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

RUN groupadd -r appuser && useradd -r -g appuser -m -d /home/appuser appuser

WORKDIR /app

COPY --chown=appuser:appuser *.py ./

RUN mkdir -p /app/model && chown -R appuser:appuser /app

ENV TRANSFORMERS_CACHE=/app/cache
ENV HUGGINGFACE_HUB_CACHE=/app/cache
ENV HF_HOME=/app/cache
ENV MODEL_PATH=/app/model

USER appuser

EXPOSE 50051
EXPOSE 5000

VOLUME /app/model

CMD ["python", "main.py"]
