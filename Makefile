all:
	go build

clean:
	rm AndoPromacUI

test:
	./AndoPromacUI

init:
	go mod init

tidy:
	go mod tidy

install:
	go build
	cp -av AndoPromacUI $(HOME)/bin