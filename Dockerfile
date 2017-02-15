# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/kudinovdenis/csServer/

# Build the csServer command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/go-sql-driver/mysql/
RUN go get github.com/jinzhu/gorm/
RUN go install github.com/kudinovdenis/csServer/

# Run the csServer command by default when the container starts.
ENTRYPOINT /go/bin/csServer

# Document that the service listens on port 80
EXPOSE 8080
