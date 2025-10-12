package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func convertValue(val cty.Value) (interface{}, error) {
	jsonBytes, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func bodyToMap(body hcl.Body, ctx *hcl.EvalContext) (map[string]interface{}, hcl.Diagnostics) {
	var allDiags hcl.Diagnostics
	result := make(map[string]interface{})

	// Try to type assert to *hclsyntax.Body to access Attributes and Blocks directly
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		// Fallback to JustAttributes if not hclsyntax.Body
		attrs, diags := body.JustAttributes()
		allDiags = append(allDiags, diags...)

		for name, attr := range attrs {
			val, valDiags := attr.Expr.Value(ctx)
			allDiags = append(allDiags, valDiags...)
			if valDiags.HasErrors() {
				continue
			}

			converted, err := convertValue(val)
			if err != nil {
				continue
			}

			result[name] = converted
		}

		return result, allDiags
	}

	// Process attributes directly from syntaxBody
	for name, attr := range syntaxBody.Attributes {
		val, valDiags := attr.Expr.Value(ctx)
		allDiags = append(allDiags, valDiags...)
		if valDiags.HasErrors() {
			continue
		}

		converted, err := convertValue(val)
		if err != nil {
			allDiags = append(allDiags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Failed to convert value",
				Detail:   err.Error(),
			})
			continue
		}

		result[name] = converted
	}

	// Process blocks directly from syntaxBody
	blocksByType := make(map[string][]interface{})
	for _, block := range syntaxBody.Blocks {
		bodyMap, bodyDiags := bodyToMap(block.Body, ctx)
		allDiags = append(allDiags, bodyDiags...)

		// Handle labels - wrap each in a map and then an array to match v1 format
		if len(block.Labels) > 0 {
			// Build nested structure for labels, wrapping each label in an array
			labelMap := bodyMap
			for i := len(block.Labels) - 1; i >= 0; i-- {
				newMap := make(map[string]interface{})
				newMap[block.Labels[i]] = []interface{}{labelMap}
				labelMap = newMap
			}
			blocksByType[block.Type] = append(blocksByType[block.Type], labelMap)
		} else {
			blocksByType[block.Type] = append(blocksByType[block.Type], bodyMap)
		}
	}

	// Add blocks to result
	for blockType, blocks := range blocksByType {
		result[blockType] = blocks
	}

	return result, allDiags
}

func main() {
	buffer, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(buffer, "stdin.hcl")
	if diags.HasErrors() {
		log.Fatal(diags.Error())
	}

	result, diags := bodyToMap(file.Body, nil)
	if diags.HasErrors() {
		log.Fatal(diags.Error())
	}

	// MarshalIndent for pretty JSON output
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}
