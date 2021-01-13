FROM golang:latest

# Add the source files
WORKDIR /go/src/github.com/sosodev/heart/
ADD build build
ADD config config
ADD kv kv
ADD las las 
ADD modules modules
ADD pool pool
COPY main.go go.mod go.sum ./

# Install LuaJIT dev libs
RUN apt update && apt install -y --no-install-recommends libluajit-5.1-dev
# Build the dynamically linked binary
RUN go build -tags luajit

ENTRYPOINT [ "./heart" ]
