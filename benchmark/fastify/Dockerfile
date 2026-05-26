FROM node:22-bookworm-slim

WORKDIR /app
COPY package.json ./
RUN npm install --omit=dev
COPY src ./src
RUN mkdir -p /app/data
ENV PORT=3002 DB_PATH=/app/data/db.sqlite SERVICE_NAME=fastify SERVICE_VERSION=2.0.0 LOG_LEVEL=silent
EXPOSE 3002
CMD ["npm", "start"]
