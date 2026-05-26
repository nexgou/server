FROM php:8.3-cli-alpine

RUN apk add --no-cache sqlite-dev \
    && docker-php-ext-install pdo_sqlite
WORKDIR /app
COPY public ./public
RUN mkdir -p /app/data
ENV PORT=3007 DB_PATH=/app/data/db.sqlite SERVICE_NAME=ajax-php SERVICE_VERSION=2.0.0 LOG_LEVEL=silent
EXPOSE 3007
CMD ["sh", "-c", "php -S 0.0.0.0:${PORT} -t public"]
