# ==========================================
# Stage 1: Build the Go Applications
# ==========================================
FROM golang:1.25.5-alpine3.23 AS go-builder
WORKDIR /build

RUN apk add --no-cache git
COPY . .

# Build binaries (cd into dir -> build -> output to /bin)
RUN cd npm-dashboard-api    && CGO_ENABLED=0 GOOS=linux go build -o /bin/npm-stats .
RUN cd charts-api           && CGO_ENABLED=0 GOOS=linux go build -o /bin/charts-api .
RUN cd avatars-api          && CGO_ENABLED=0 GOOS=linux go build -o /bin/avatars-api .
RUN cd locations-api        && CGO_ENABLED=0 GOOS=linux go build -o /bin/locations-api .
RUN cd institutions-api     && CGO_ENABLED=0 GOOS=linux go build -o /bin/institutions-api .
RUN cd github-dashboard-api && CGO_ENABLED=0 GOOS=linux go build -o /bin/github-stats .

# ==========================================
# Stage 2: Build the TypeScript Application
# ==========================================
FROM node:20-alpine AS node-builder
WORKDIR /app/calendars-api

COPY calendars-api/package*.json ./
RUN npm install

COPY calendars-api/ .
RUN npm run build
RUN npm prune --production

# ==========================================
# Stage 3: The Final Production Image
# ==========================================
FROM alpine:latest

RUN apk add --no-cache bash nodejs ca-certificates rsvg-convert ttf-dejavu

COPY --from=caddy:2-alpine /usr/bin/caddy /usr/bin/caddy

WORKDIR /app

# Copy Go Binaries
COPY --from=go-builder /bin/github-stats ./bin/
COPY --from=go-builder /bin/npm-stats ./bin/
COPY --from=go-builder /bin/charts-api ./bin/
COPY --from=go-builder /bin/avatars-api ./bin/
COPY --from=go-builder /bin/locations-api ./bin/
COPY --from=go-builder /bin/institutions-api ./bin/

# Copy Node App
COPY --from=node-builder /app/calendars-api/dist ./calendars-api/dist
COPY --from=node-builder /app/calendars-api/node_modules ./calendars-api/node_modules
COPY --from=node-builder /app/calendars-api/package.json ./calendars-api/

# Copy Static Assets
COPY public/ ./public/
COPY Caddyfile .
COPY start.sh .

RUN chmod +x start.sh
EXPOSE 7999
CMD ["./start.sh"]