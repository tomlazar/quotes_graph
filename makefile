run:
	go build
	docker-compose up -d --build

build:
	docker-compose build --force-rm --no-cache