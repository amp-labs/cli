# Generate Go types from OpenAPI spec.

MANIFEST_YAML = https://raw.githubusercontent.com/amp-labs/openapi/main/manifest/manifest.yaml
PROBLEM_YAML = https://raw.githubusercontent.com/amp-labs/openapi/main/problem/problem.yaml

.PHONY: gen/manifest
gen/manifest:
	oapi-codegen --config=oapi.config.yaml --o=manifest.gen.go $(MANIFEST_YAML)

.PHONY: gen/problem
gen/problem:
	oapi-codegen --config=oapi.config.yaml --o=problem.gen.go $(PROBLEM_YAML)
