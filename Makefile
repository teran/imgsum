export GOPATH := $(PWD)
export GOBIN := $(GOPATH)/bin
export PACKAGES := $(shell env GOPATH=$(GOPATH) go list ./src/imgsum/...)

all: clean predependencies dependencies build

clean:
	rm -vf bin/*

build: build-macos build-linux build-windows

build-macos: build-macos-amd64

build-linux: build-linux-amd64

build-windows: build-windows-amd64

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-darwin-amd64 imgsum/cmd

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-linux-amd64 imgsum/cmd

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-windows-amd64.exe imgsum/cmd

dependencies:
	cd src && trash

predependencies:
	go get -u github.com/rancher/trash

sign:
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-darwin-amd64.sig 			bin/imgsum-darwin-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-linux-amd64.sig 				bin/imgsum-linux-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-windows-amd64.exe.sig 	bin/imgsum-windows-amd64.exe

verify:
	gpg --verify bin/imgsum-darwin-amd64.sig 			bin/imgsum-darwin-amd64
	gpg --verify bin/imgsum-linux-amd64.sig 				bin/imgsum-linux-amd64
	gpg --verify bin/imgsum-windows-amd64.exe.sig 	bin/imgsum-windows-amd64.exe
