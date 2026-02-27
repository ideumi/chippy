export GO_BUILD = CGO_ENABLED=0 go build -trimpath

build:
	$(GO_BUILD) -o chippy cmd/chippy/*.go

	./chippy check all installer/
	./chippy check all lib/
	./chippy check all misc/

release:
	@mkdir -p rel
	@mkdir -p installer/out
	@cd rel && ../scripts/make-release.sh

release-all:
	@mkdir -p rel
	@mkdir -p installer/out
	@cd rel && ../scripts/release-all.sh
