all: build/share/man/man1/cgen.1

build:
	@mkdir -p build build-arm64 build-amd64

build/cgen build/docgen: build $(shell find . -type f -name "*.go")
	GOOS=linux GOARCH=arm64	CGO_ENABLED=0 go build -o build-arm64 -ldflags="-s -w" ./...
	GOOS=linux GOARCH=amd64	CGO_ENABLED=0 go build -o build-amd64 -ldflags="-s -w" ./...
	CGO_ENABLED=0 go build -o build -ldflags="-s -w" ./...

build/share/man/man1/cgen.1: build/docgen
	@build/docgen

clean:
	@rm -rf build*

pack: build/share/man/man1/cgen.1
	ARCH=amd64 DIST_TAG=.fc43 envsubst < nfpm.yml | nfpm pkg -f - --packager rpm --target build-amd64/
	ARCH=amd64 DIST_TAG=.fc44 envsubst < nfpm.yml | nfpm pkg -f - --packager rpm --target build-amd64/
	ARCH=amd64 DIST_TAG=      envsubst < nfpm.yml | nfpm pkg -f - --packager deb --target build-amd64/
	ARCH=arm64 DIST_TAG=.fc43 envsubst < nfpm.yml | nfpm pkg -f - --packager rpm --target build-arm64/
	ARCH=arm64 DIST_TAG=.fc44 envsubst < nfpm.yml | nfpm pkg -f - --packager rpm --target build-arm64/
	ARCH=arm64 DIST_TAG=      envsubst < nfpm.yml | nfpm pkg -f - --packager deb --target build-arm64/
