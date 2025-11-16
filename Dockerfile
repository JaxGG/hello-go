# 构建阶段
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o hello .

# 运行阶段
FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/hello .
EXPOSE 8080
CMD ["./hello"]
