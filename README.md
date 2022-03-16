# JSON Schema Validator

[![Go Report Card](https://goreportcard.com/badge/github.com/earthrise-media/schema-validator-action)](https://goreportcard.com/report/github.com/earthrise-media/schema-validator-action)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fearthrise-media%2Fschema-validator-action.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fearthrise-media%2Fschema-validator-action?ref=badge_shield)

This is a utility (and Github action) that recursively walks a directory and validates all JSON files that it finds.
Based on [this great validation library](https://github.com/santhosh-tekuri/jsonschema)


## Usage 

The `schema-validator` looks for the following environment variables

- `GITHUB_WORKSPACE` the directory to walk, all subdirectories are also inspected. 
When run as an action github will populate this value with the root of the repository  
- `FORCE_SCHEMA_LOCATION` (optional) the location of a schema to use for validation(can be file path or http/s) 
- `FAIL_FAST` (default `false`) if set the tool will exit on first error 
- `REQUIRE_SCHEMAS` (default `false`) setting this will cause JSON files without declared schemas to be considered validation failures


### Examples 

Make sure every JSON file on your machine is GeoJSON, fail if one isn't:

`$ env FAIL_FAST=true REQUIRE_SCHEMAS=https://json.schemastore.org/geojson.json GITHUB_WORKSPACE=/`schema-validator`

Inside a GitHub workflow:
```
steps:
   - uses: actions/checkout@v2
   - uses: earthrise-media/schema-validator-action@main
```

## Behavior

Files are validated only if they end in `.json` or `.geojson`
Validation includes:
1. Must be valid JSON syntax (i.e. braces and quotes must be closed etc.)
2. Schema validation:
   - If a schema is provided to the tool (using the `FORCE_SCHEMA_LOCATION` env var) it will override any schema declared in the JSON file
   - If no schema is provided the file will be validated using any schema declared in a top level `$schema` field
   - If no schema is found in the file, it will be considered valid unless the `REQUIRE_SCHEMAS` env var is set, in which case it will be considered a failure

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fearthrise-media%2Fschema-validator-action.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fearthrise-media%2Fschema-validator-action?ref=badge_large)
