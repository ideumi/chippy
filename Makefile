build:
	go build -o chippy cmd/chippy/*.go

	./chippy check all installer/
	./chippy check all lib/
	./chippy check all misc/

release:
	@mkdir -p rel
	@mkdir -p installer/out
	@cd rel && ../scripts/make-release.sh
