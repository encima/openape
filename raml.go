package openape

import (
	"fmt"

	"github.com/Jumpscale/go-raml/raml"
)

// MapRAMLModels iterates the types specified in a raml file
func (oape *OpenApe) MapRAMLModels() {
	for k, v := range oape.ramlAPI.Types {
		oape.db.CreateRAMLSchema(k, v)
	}
}

// MapRAMLResources iterates the types specified in a raml file
func (oape *OpenApe) MapRAMLResources(res map[string]*raml.Resource) {
	for k, resVal := range res {
		model := resVal.Type.Name
		fmt.Println(k)
		for _, methodVal := range resVal.Methods {
			oape.AddRoute(resVal.URI, methodVal.Name, model)
		}
		if len(resVal.Nested) > 0 {
			oape.MapRAMLResources(resVal.Nested)
		}
	}
}
