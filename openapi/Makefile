# Generate Go types from OpenAPI spec.
.PHONY: all
all:
	echo "No default target."

.PHONY: gen
gen:
	oapi-codegen --config=oapi.config.yaml --o=manifest.gen.go manifest.yaml
	oapi-codegen --config=oapi.config.yaml --o=problem.gen.go problem.yaml
