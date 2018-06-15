SHELL = /bin/bash

dvmapp: cmd/dvmapp/main.go
	go build -o $@ $<

clean:
	rm -f dvmapp

data.db:
	sqlite3 $@ < db.sql
