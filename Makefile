gogetter: gogetter.go
	go build

run:
	go run gogetter.go

deploy:
	git aws.push

clean:
	rm gogetter

dev: gogetter
	PROTOCOL=http gin -b gogetter

.PHONY: deploy run clean
