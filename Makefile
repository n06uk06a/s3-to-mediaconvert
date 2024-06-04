.PHONY: build

build:
	GOOS=linux go build -o build/s3_to_mediaconvert
