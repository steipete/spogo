.PHONY: spogo docs-site

spogo:
	go build -o spogo ./cmd/spogo

docs-site:
	@node scripts/build-docs-site.mjs
