SHELL = /bin/bash

dvmapp: cmd/dvmapp/main.go
	go generate cmd/dvmapp/main.go
	go build -o $@ $<

clean:
	rm -f dvmapp
