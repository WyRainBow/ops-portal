FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ops-portal-api ./apps/api

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata git nodejs npm
# pm2 CLI is used by /api/admin/runtime/status to read host PM2 daemon state.
# With PM2_HOME mounted from host, pm2 in container can query host processes.
RUN npm i -g pm2
COPY --from=build /out/ops-portal-api /app/ops-portal-api

ENV OPS_PORTAL_API_PORT=18081
EXPOSE 18081
CMD ["/app/ops-portal-api"]
