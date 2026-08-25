test:
	go test

build:
	go build -o ./plata-go-assignment .

run: build
	./plata-go-assignment

e2e:
	node ./minie2e.js
