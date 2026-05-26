FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=1 GOOS=linux go build -o /cloudalbum .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /cloudalbum ./cloudalbum
COPY configs ./configs
RUN mkdir -p /app/data/images
EXPOSE 8080
VOLUME ["/app/data"]
CMD ["./cloudalbum"]
