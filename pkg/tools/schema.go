package tools

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

// GenerateSchema creates a JSON schema for the given type
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}