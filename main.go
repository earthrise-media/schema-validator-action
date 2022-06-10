package main

import (
	"encoding/json"
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
var cachedSchemas = make(map[string]*jsonschema.Schema)
var schemaErrors = make(map[string]error)
var hadError bool

const (
	DIR                 = "GITHUB_WORKSPACE"
	ForceSchemaLocation = "FORCE_SCHEMA_LOCATION"
	FailFast            = "FAIL_FAST"
	RequireSchemas      = "REQUIRE_SCHEMAS"
)

func main() {

	viper.SetDefault(DIR, "")
	viper.SetDefault(ForceSchemaLocation, "")
	viper.SetDefault(FailFast, false)
	viper.SetDefault(RequireSchemas, false)

	viper.AutomaticEnv()

	var err error

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	if viper.GetString(ForceSchemaLocation) != "" {
		compiledSchema, err = compiler.Compile(viper.GetString(ForceSchemaLocation))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to compile provided schema: %#v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s is a required ENV", ForceSchemaLocation)
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
	if err != nil && viper.GetBool(FailFast) {
		fmt.Println(fmt.Sprintf("Validation failed fast, some JSON files were potentially skipped!"))
	}

	for k := range schemaErrors {
		schemaError := schemaErrors[k]
		if schemaError == nil {
			fmt.Println(fmt.Sprintf("%s \U00002705", k))

		} else {
			fmt.Println(fmt.Sprintf("%s \U0000274C", k))
			fmt.Println(fmt.Sprintf("Error detail: %s", schemaError.Error()))
		}
	}
	if hadError {
		os.Exit(1)
	}

}

func walkValidate(entry string, dir fs.DirEntry, err error) error {

	if dir != nil {
		if dir.IsDir() {
			return nil
		}
	}
	if strings.HasSuffix(entry, "json") {
		fmt.Println(fmt.Sprintf("Validating %s", entry))
		err = validate(entry)
		schemaErrors[entry] = err

		if err != nil {
			hadError = true
			if viper.GetBool(FailFast) {
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
		return fmt.Errorf("Error opening file: %v\n", err)
	}

	//var v map[string]interface{}
	var v interface{}
	dec := json.NewDecoder(file)
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("Syntax error: %v\n", err)
	}

	err = compiledSchema.Validate(v)
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		out := ve.DetailedOutput()
		b, _ := json.MarshalIndent(out, "", "  ")
		return fmt.Errorf(string(b))
	}
	return nil

}
