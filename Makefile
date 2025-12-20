build:
	go build -o chippy cmd/chippy/*.go

	./chippy check all installer/
	./chippy check all lib/
	./chippy check all misc/
	./chippy check all plugins/example

	@for dir in plugins/*/; do \
		if [ -f "$$dir/Makefile" ] && [ "$$(basename $$dir)" != "example" ]; then \
			(cd "$$dir" && $(MAKE) build) || exit 1; \
		fi; \
	done

release:
	@mkdir -p rel
	@mkdir -p installer/out
	@cd rel && ../scripts/make-release.sh
