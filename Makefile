CGO_CFLAGS="-I${PWD}/gattlib/include"
CGO_LDFLAGS="-L${PWD}/gattlib/build/dbus"
GF=CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} LD_LIBRARY_PATH=${PWD}/gattlib/build/dbus

all: gattlib test build

test:
	${GF} go test -v $$(go list ./... | grep -v /vendor/)

build:
	${GF} go build -o influx-cli ./influx

gattlib:
	git clone https://github.com/labapart/gattlib
	mkdir gattlib/build
	cd gattlib/build && cmake -DGATTLIB_FORCE_DBUS=TRUE .. && make


gattlib-clean:
	rm -rf ./gattlib

clean: gattlib-clean
	rm -f ./influx-cli

