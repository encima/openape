package openape

// MapRAMLModels iterates the types specified in a raml file
func (oape *OpenApe) MapRAMLModels() {
	for k, v := range oape.ramlAPI.Types {
		oape.db.CreateRAMLSchema(k, v)
	}
}

// MapRAMLResources iterates the types specified in a raml file
func (oape *OpenApe) MapRAMLResources() {
	for _, resVal := range oape.ramlAPI.Resources {
		model := resVal.Type.Name
		for _, methodVal := range resVal.Methods {
			oape.AddRoute(resVal.URI, methodVal.Name, model)
		}
	}
}
