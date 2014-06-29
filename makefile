all: local

local:
	go build

beagleboard:
	GOARM=7 GOARCH=arm GOOS=linux go build

install:
	go install
	GOARM=7 GOARCH=arm GOOS=linux go install
