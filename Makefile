export GOPATH := $(PWD)
export GOBIN := $(GOPATH)/bin
export PACKAGES := $(shell env GOPATH=$(GOPATH) go list ./src/imgsum/...)

all: clean predependencies dependencies build

clean:
	rm -vf bin/*

build: build-macos build-linux build-windows

build-macos: build-macos-amd64 build-macos-i386

build-linux: build-linux-amd64 build-linux-i386

build-windows: build-windows-amd64 build-windows-i386

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-darwin-amd64 imgsum/cmd

build-macos-i386:
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -o bin/imgsum-darwin-i386 imgsum/cmd

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-linux-amd64 imgsum/cmd

build-linux-i386:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o bin/imgsum-linux-i386 imgsum/cmd

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/imgsum-windows-amd64.exe imgsum/cmd

build-windows-i386:
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o bin/imgsum-windows-i386.exe imgsum/cmd

dependencies:
	cd src && trash

predependencies:
	go get -u github.com/rancher/trash

sign:
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-darwin-amd64.sig 				bin/imgsum-darwin-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-darwin-i386.sig 				bin/imgsum-darwin-i386
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-linux-amd64.sig 				bin/imgsum-linux-amd64
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-linux-i386.sig 					bin/imgsum-linux-i386
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-windows-amd64.exe.sig 	bin/imgsum-windows-amd64.exe
	gpg --detach-sign --digest-algo SHA512 --no-tty --batch --output bin/imgsum-windows-i386.exe.sig 		bin/imgsum-windows-i386.exe

verify:
	gpg --verify bin/imgsum-darwin-amd64.sig 				bin/imgsum-darwin-amd64
	gpg --verify bin/imgsum-darwin-i386.sig 				bin/imgsum-darwin-i386
	gpg --verify bin/imgsum-linux-amd64.sig 				bin/imgsum-linux-amd64
	gpg --verify bin/imgsum-linux-i386.sig 					bin/imgsum-linux-i386
	gpg --verify bin/imgsum-windows-amd64.exe.sig 	bin/imgsum-windows-amd64.exe
	gpg --verify bin/imgsum-windows-i386.exe.sig 		bin/imgsum-windows-i386.exe
