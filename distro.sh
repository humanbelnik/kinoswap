#!/bin/bash


N=${1:-500}
URL="http://localhost:80/api/v1/movies"

count_8080=0
count_8888=0
count_8887=0

for ((i=1; i<=N; i++)); do
    response=$(curl -s -i -X 'GET' \
        -H 'accept: application/json' -H 'x-admin-token: 66d94e0f-6091-4d30-b90a-6b4da1abda47' \
        "$URL" 2>/dev/null)
    
    upstream_header=$(echo "$response" | grep -i "x-upstream" | head -1)
    if [ -n "$upstream_header" ]; then
        port=$(echo "$upstream_header" | grep -oE ':[0-9]{4}' | tail -1 | cut -c2-5)
        case "$port" in
            "8080")
                ((count_8080++))
                ;;
            "8888")
                ((count_8888++))
                ;;
            "8887")
                ((count_8887++))
                ;;
            *)
                ((count_unknown++))
                echo "Request $i: Unknown port - $upstream_header"
                ;;
        esac
    else
        http_code=$(echo "$response" | grep -i "HTTP/" | head -1 | awk '{print $2}')
        if [ -n "$http_code" ]; then
            ((count_unknown++))
            echo "Request $i: No X-Upstream header (HTTP $http_code)"
        else
            ((count_error++))
            echo "Request $i: Connection error"
        fi
    fi
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "Processed $i/$N requests..."
    fi
done

echo "8080: $count_8080 ($(echo "scale=1; $count_8080 * 100 / $N" | bc)%)"
echo "8888: $count_8888 ($(echo "scale=1; $count_8888 * 100 / $N" | bc)%)"
echo "8887: $count_8887 ($(echo "scale=1; $count_8887 * 100 / $N" | bc)%)"
echo "SUM $N requests"


