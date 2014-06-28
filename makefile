all: local

local:
	go install

beagleboard:
	GOARM=7 GOARCH=arm GOOS=linux go install
