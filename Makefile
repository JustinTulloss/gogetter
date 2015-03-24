IMG=jmtulloss/gogetter

gogetter: gogetter.go
	go build -o gogetter cmd/main.go

gogetter-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gogetter cmd/main.go

docker-image: gogetter-linux
	docker build -t $(IMG) .

push: docker-image
	docker push $(IMG)

run:
	go run cmd/main.go

clean:
	rm gogetter

.PHONY: run clean push docker-image gogetter-linux
