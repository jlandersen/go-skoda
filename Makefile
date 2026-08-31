OAPI_CODEGEN_VERSION := v2.8.0
GO_IMAGE := golang:1.26.1@sha256:cd78d88e00afadbedd272f977d375a6247455f3a4b1178f8ae8bbcb201743a8a
OPENAPI_URL := https://public.api.connect.skoda-auto.cz/v3/api-docs

.PHONY: generate update-openapi
generate:
	docker run --rm \
		--user "$$(id -u):$$(id -g)" \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/tmp/go-mod-cache \
		-e GOPATH=/tmp/go \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		$(GO_IMAGE) \
		go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION) \
		-config oapi-codegen.yaml api/openapi.json

update-openapi:
	curl -fsSL "$(OPENAPI_URL)" -o api/openapi.json
	$(MAKE) generate
