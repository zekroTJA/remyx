FROM golang:alpine AS build-be
WORKDIR /build
COPY cmd/ cmd/
COPY internal/ internal/
COPY go.mod .
COPY go.sum .
RUN go build -o remyx cmd/remyx/main.go

FROM node:alpine AS build-fe
WORKDIR /build
COPY web/ .
RUN yarn
RUN yarn build

FROM alpine AS final
WORKDIR /app
COPY --from=build-be /build/remyx .
COPY --from=build-fe /build/dist web/dist
COPY migrations migrations
EXPOSE 80
ENV REMYX_WEBSERVER_BINDADDRESS=":80"
ENTRYPOINT [ "/app/remyx" ]