.PHONY: all
all: clean build run

build:
	mkdir build
	go build -o build/server main.go
	npx tailwindcss --minify -c tailwind.config.js -i css/index.css -o build/static/css/index.css
	cp -R templates build/templates
	cp -R images build/static/images

.PHONY: run
run: build
	cd build && ./server

.PHONY: clean
clean:
	rm -rf build
