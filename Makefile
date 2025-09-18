build:
	go build -o chippy cmd/chippy/*.go

release:
	@mkdir -p rel
	@mkdir -p installer/out
	@mkdir -p misc/avant/out
	@cd rel && ../scripts/make-release.sh
