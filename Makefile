.PHONY: all build run
all: build run
build: 
	mkdir -p build; cd build; go build ..
run:
	./build/k8svirt
