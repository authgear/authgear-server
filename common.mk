DIST ?= skygear

.PHONY: build
build:
	go build -o $(DIST)
	chmod +x $(DIST)

.PHONY: clean
clean:
	-rm $(DIST)

