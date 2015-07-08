
all: run


build:
	go build -o gank.exe

run: build
	./gank.exe > out.log 2>&1 & disown

