all: local

local:
	go build

bb:
	GOARM=7 GOARCH=arm GOOS=linux go build

install:
	go install

install_bb:
	GOARM=7 GOARCH=arm GOOS=linux go install

clean:
	rm ../../../../bin/gowdc

clean_bb:
	rm ../../../../bin/linux_arm/gowdc
