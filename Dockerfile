# 生产多阶段构建 Dockerfile

# Stage 1: 构建后端
FROM golang:1.23-alpine AS backend-builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /drill-server ./cmd/server

# Stage 2: 构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /web

COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install

COPY web/ .

RUN npm run build

# Stage 3: 运行时镜像
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /drill-server .
COPY --from=frontend-builder /web/dist /app/web/dist
COPY configs/config.yaml /app/configs/config.yaml
COPY scripts/init-db.sql /app/scripts/init-db.sql

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./drill-server"]
