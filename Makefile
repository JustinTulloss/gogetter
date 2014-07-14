gogetter: gogetter.go
	go build

run:
	go run gogetter.go

deploy:
	git aws.push

clean:
	rm gogetter

.PHONY: deploy run clean
