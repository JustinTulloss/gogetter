IMG=jmtulloss/gogetter

gogetter: gogetter.go
	go build

gogetter-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

docker-image: gogetter-linux
	docker build -t $(IMG) .

push: docker-image
	docker push $(IMG)

run:
	go run gogetter.go

deploy:
	git aws.push

clean:
	rm gogetter

dev: gogetter
	PROTOCOL=http gin -b gogetter

.PHONY: deploy run clean
