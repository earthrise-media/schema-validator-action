package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
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

func usage() {
	flag.PrintDefaults()
}

func main() {

	dir := flag.String("dir", "/", "directory to search for json files")
	// draft := flag.Int("draft", 2020, "which schema draft version to use")
	schema := flag.String("schema", "", "the schema to apply, if empty will use schema references in json docs")
	failFast = flag.Bool("failFast", false, "setting this to true will cause the validator to exit on the first failure")
	requireSchema = flag.Bool("requireSchema", false, "setting this will cause json without declared schemas to be considered validation failures")

	flag.Usage = usage
	flag.Parse()
	//if len(flag.Args()) == 0 {
	//	usage()
	//	os.Exit(1)
	//}
	var err error

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	if schema != nil && *schema != "" {
		compiledSchema, err = compiler.Compile(*schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unabled to compile provided schema: %#v\n", err)
			os.Exit(1)
		}
	}

	err = filepath.WalkDir(*dir, walkValidate)
	if err != nil && *failFast {
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

	if dir.IsDir() {
		return nil
	}
	if strings.HasSuffix(entry, ".json") || strings.HasSuffix(entry, ".geojson") {

		err = validate(entry)
		schemaErrors[entry] = err

		if err != nil {
			fmt.Println(err.Error())
			if *failFast {
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
			} else {
				//schema field found but empty
				if *requireSchema {
					return errors.New(fmt.Sprintf("empty schema declaration found in %s and requireSchema is set", jsonFile))
				} else {
					return nil
				}
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
