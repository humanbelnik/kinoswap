FROM nginx:1.24-alpine

RUN apk add --no-cache \
    nodejs \
    npm \
    && npm install -g marked

COPY nginx.conf /etc/nginx/nginx.conf
COPY conf.d /etc/nginx/conf.d

RUN mkdir -p /usr/share/nginx/static

RUN echo "worker_processes auto;" > /etc/nginx/nginx.conf.part

RUN echo '#!/bin/sh' > /etc/nginx/start.sh \
    && echo 'nginx -g "daemon off;"' >> /etc/nginx/start.sh \
    && chmod +x /etc/nginx/start.sh

EXPOSE 80 443

CMD ["/etc/nginx/start.sh"]