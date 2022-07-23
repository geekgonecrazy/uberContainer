FROM node:12-buster AS frontend

WORKDIR /app

COPY web/package.json .
COPY web/package-lock.json .

RUN npm install

COPY web .

RUN npm run build

FROM golang:1.17-alpine AS backend

RUN apk add --no-cache ca-certificates git
WORKDIR /go/src/github.com/FideTechSolutions/uberContainer/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine as runtime

RUN apk add --no-cache imagemagick

WORKDIR /app

ENV GIN_MODE=release

COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=frontend /app/public web/public
COPY --from=backend /go/src/github.com/FideTechSolutions/uberContainer/uberContainer uberContainer
COPY --from=backend /go/src/github.com/FideTechSolutions/uberContainer/templates templates

EXPOSE 8080

CMD ["./uberContainer"]