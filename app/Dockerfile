FROM golang:alpine as build

WORKDIR /app

COPY . .

RUN go build .

FROM golang:alpine

COPY --from=build /app/observability-demo-app /app/observability-demo-app

ENTRYPOINT [ "/app/observability-demo-app" ]
