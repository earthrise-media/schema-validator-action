# JSON Schema Validator
This is a utility (and Github action) that recursively walks a directory and validates all JSON files that it finds.
Based on [this great validation library](github.com/santhosh-tekuri/jsonschema)


## Usage 

The `schema-validator` looks for the following environment variables

- `dir` the directory to walk, all subdirectories are also inspected
- `schema` (optional) the location of a schema to use for validation(can be file path or http/s) 
- `failFast` (default `false`) if set the tool will exit on first error 
- `requireSchema` (default `false`) setting this will cause JSON files without declared schemas to be considered validation failures


### Examples 

Make sure every JSON file on your machine is GeoJSON, fail if one isn't:

`$ schema-validator -failfast=true -schema=https://json.schemastore.org/geojson.json -dir=/`


## Behavior

Files are validated only if they end in `.json` or `.geojson`
Validation includes:
1. Must be valid JSON syntax (i.e. braces and quotes must be closed etc.)
2. Schema validation:
   - If a schema is provided to the tool (using the `-schema` flag) it will override any schema declared in the JSON file
   - If no schema is provided the file will be validated using any schema declared in a top level `$schema` field
   - If no schema is found either way 