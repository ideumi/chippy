build:
	go build -o chippy cmd/chippy/*.go

release:
	@mkdir -p rel
	@cd rel && ../scripts/make-release.sh
