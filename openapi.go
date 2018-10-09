package openape

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// MapModels reads the models from the provided swagger file and creates the correspdonding tables in Postgres
func (oape *OpenApe) MapModels(models map[string]*openapi3.SchemaRef) {
	for k, v := range models {
		oape.db.CreateSchema(k, v.Value.Properties)
	}
}

// GetModelFromPath identifies which routes maps to which models identified in the Schemas of the spec
func (oape *OpenApe) GetModelFromPath(path string) string {
	for k := range oape.swagger.Components.Schemas {
		if strings.Contains(strings.ToLower(path), strings.ToLower(k)) {
			return k
		}
	}
	return ""
}

// MapRoutes iterates the paths laid out in the swagger file and adds them to the router
func (oape *OpenApe) MapRoutes(paths map[string]*openapi3.PathItem) {
	for k, v := range paths {
		// TODO handle when user specifies function and do not pass to route
		model := oape.GetModelFromPath(k)
		for opName := range v.Operations() {
			oape.AddRoute(k, opName, model)
		}
	}
}
