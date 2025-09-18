build:
	go build -o chippy cmd/chippy/*.go

release:
	@mkdir -p rel
	@mkdir -p installer/out
	@cd rel && ../scripts/make-release.sh
