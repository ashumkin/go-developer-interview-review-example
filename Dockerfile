FROM golang:1.22-alpine3.18

WORKDIR /go/src/app
COPY . ./
RUN go mod download
RUN go mod tidy && \
    go build -o ./build/server .

WORKDIR /srv
RUN cp /go/src/app/build/server .

CMD [ "/srv/server" ]
