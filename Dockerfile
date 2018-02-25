# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/yamamushi/TheAether

# Create our shared volume
RUN mkdir /AetherData

# Get the du-discordbot dependencies inside the container.
RUN cd /go/src/github.com/yamamushi/TheAether && go get ./...

# Install and run TheAether
RUN go install github.com/yamamushi/TheAether

# Run the command by default when the container starts.
WORKDIR /AetherData
ENTRYPOINT /go/bin/TheAether

VOLUME /AetherData