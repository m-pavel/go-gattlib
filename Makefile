CGO_CFLAGS="-I${PWD}/gattlib/include"
CGO_LDFLAGS="-L${PWD}/gattlib/build/dbus"
GF=CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} LD_LIBRARY_PATH=${PWD}/gattlib/build/dbus

all: gattlib test build

deps:
	${GF} go get -v -d ./...
test: deps
	${GF} go test -v $$(go list ./... | grep -v /vendor/)

build: deps
	${GF} go build -work -o tion-influx-cli ./influx
	${GF} go build -o tion-schedule ./schedule
	${GF} go build -o cli ./tion-cli 

gattlib:
	git clone https://github.com/labapart/gattlib
	mkdir gattlib/build
	cd gattlib/build && cmake -DGATTLIB_FORCE_DBUS=TRUE .. && make


gattlib-clean:
	rm -rf ./gattlib

clean: gattlib-clean
	rm -f ./tion-influx-cli
	rm -f ./tion-schedule
	rm -f ./tion-cli


