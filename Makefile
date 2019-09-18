CGO_CFLAGS="-I${PWD}/gattlib/include"
CGO_LDFLAGS="-L${PWD}/gattlib/build/dbus -lm -lutil"
GF=CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} LD_LIBRARY_PATH=${PWD}/gattlib/build/dbus

all: gattlib test

deps:
	${GF} go get -v -d ./...
test: deps
	${GF} go test -v $$(go list ./... | grep -v /vendor/)

gattlib:
	git clone https://github.com/labapart/gattlib
	mkdir gattlib/build
	cd gattlib/build && cmake -DGATTLIB_FORCE_DBUS=TRUE .. && make

gattlib-clean:
	rm -rf ./gattlib

clean: gattlib-clean
