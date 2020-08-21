FROM golang:alpine as builder

WORKDIR /app

COPY . .

ENV GO111MODULE=on
# statically link our binary so we can pop it into an empty container CGO_ENABLED=0
#
# and of course my first example... can't run CGO_ENABLED=0 because sqlite breaks
# still try with 0 first
# only switch to this if you need MOAR cross-compilation of your static binary
RUN apk add build-base
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o goservice /app/cmd/goservice/serv.go
RUN go vet ./...
RUN go test ./...

# we'd probably move the binaries from here off into an EMPTIER container image
# then push that image to the container registry
#FROM ubuntu:bionic-20200112 
FROM scratch

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/goservice .

# in a RW app, we'd load the config in via configmap into kubernetes or otherwise inject 
# maybe at /etc/config.json
COPY --from=builder /app/config.json .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./goservice"] 
