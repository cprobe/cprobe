BUILD_VERSION=0.0.1
BUILDFLAGS=CGO_ENABLED=0
BUILD_TIME=$(shell date +"%Y-%m-%d-%H-%M-%S")
LDFLAGS="-w -s -X 'github.com/cprobe/cprobe/lib/buildinfo.Version=$(BUILD_VERSION)-$(BUILD_TIME)'"

all: build

tidy: goenv
	go mod tidy

goenv:
	export GOPROXY=https://goproxy.cn,direct
	export GOSUMDB=off
	export GO111MODULE=on

build: tidy
	$(BUILDFLAGS) go build -ldflags $(LDFLAGS) -o cprobe .

nohup:
	nohup ./cprobe > stdout.log 2>&1 &
