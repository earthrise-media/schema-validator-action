package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var compiledSchema *jsonschema.Schema
var failFast *bool
var requireSchema *bool
var cachedSchemas = make(map[string]*jsonschema.Schema)
var schemaErrors = make(map[string]error)

const (
	DIR                   = "GITHUB_WORKSPACE"
	FORCE_SCHEMA_LOCATION = "FORCE_SCHEMA_LOCATION"
	FAIL_FAST             = "FAIL_FAST"
	REQUIRE_SCHEMAS       = "REQUIRE_SCHEMAS"
)

func main() {

	viper.SetDefault(DIR, "")
	viper.SetDefault(FORCE_SCHEMA_LOCATION, "")
	viper.SetDefault(FAIL_FAST, false)
	viper.SetDefault(REQUIRE_SCHEMAS, false)

	viper.AutomaticEnv()

	var err error

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	if viper.GetString(FORCE_SCHEMA_LOCATION) != "" {
		compiledSchema, err = compiler.Compile(viper.GetString(FORCE_SCHEMA_LOCATION))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unabled to compile provided schema: %#v\n", err)
			os.Exit(1)
		}
	}
	dir := viper.GetString(DIR)
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println(fmt.Sprintf("unable to find current working dir and no dir provided: %s", err.Error()))
			os.Exit(1)
		}
	}
	err = filepath.WalkDir(dir, walkValidate)
	if err != nil && viper.GetBool(FAIL_FAST) {
		fmt.Println(fmt.Sprintf("Validation failed fast, some JSON files were potentially skipped!"))
	}

	for k, _ := range schemaErrors {
		schemaError := schemaErrors[k]
		if schemaError == nil {
			fmt.Println(fmt.Sprintf("%s \U00002705", k))

		} else {
			fmt.Println(fmt.Sprintf("%s \U0000274C", k))
			fmt.Println(fmt.Sprintf("Error detail: %s", schemaError.Error()))
		}
	}

}

func walkValidate(entry string, dir fs.DirEntry, err error) error {

	if dir != nil {
		if dir.IsDir() {
			return nil
		}
	}
	if strings.HasSuffix(entry, ".json") || strings.HasSuffix(entry, ".geojson") {
		fmt.Println(fmt.Sprintf("Validating %s", entry))
		err = validate(entry)
		schemaErrors[entry] = err

		if err != nil {
			if viper.GetBool(FAIL_FAST) {
				return err
			}
		}
	}
	return nil
}

func validate(jsonFile string) error {

	file, err := os.Open(jsonFile)
	defer file.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("Error opening file: %v\n", err))
	}

	var v map[string]interface{}
	dec := json.NewDecoder(file)
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return errors.New(fmt.Sprintf("Syntax error: %v\n", err))
	}
	var currentSchema *jsonschema.Schema

	// this means we are forcing a schema, not looking for a declared one
	if compiledSchema != nil {
		currentSchema = compiledSchema
	} else {
		//look for schema declaration in the file
		if val, ok := v["$schema"]; ok {
			//found schema
			declaredSchema := fmt.Sprintf("%v", val)
			if declaredSchema != "" {
				currentSchema, err = loadSchema(declaredSchema)
				if err != nil {
					return err
				}
			} else {
				//schema field found but empty
				if viper.GetBool(REQUIRE_SCHEMAS) {
					return errors.New(fmt.Sprintf("empty schema declaration found in %s and requireSchema is set", jsonFile))
				} else {
					return nil
				}
			}
		} else {
			//no schema found
			if viper.GetBool(REQUIRE_SCHEMAS) {
				return errors.New(fmt.Sprintf("no schema reference found in %s and requireSchema is set", jsonFile))
			} else {
				return nil
			}
		}
	}

	err = currentSchema.Validate(v)
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		out := ve.DetailedOutput()
		b, _ := json.MarshalIndent(out, "", "  ")
		return errors.New(string(b))
	}
	return nil

}

func loadSchema(declaredSchema string) (*jsonschema.Schema, error) {

	if schema, ok := cachedSchemas[declaredSchema]; ok {
		//found it in cache
		return schema, nil
	}

	schema, err := jsonschema.Compile(declaredSchema)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to load declared schema: %v\n", err))
	} else {
		//cache it
		cachedSchemas[declaredSchema] = schema
		return schema, err
	}

}
