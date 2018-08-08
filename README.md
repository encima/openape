# OpenAPE

## WHAT?

OpenAPE is designed to be a server extension of the `OpenAPI` specification, reading in a `Swagger` file and creating routes with the specified security.

This already exists, you say? Why yes, yes it does. Many exist for OpenAPI 2 but few support OpenAPI3 and all of them require code to be written...

## WHY?

For the past 3 years, I have worked on a number of different projects that have required creation of an API, `Swagger` made this much easier for the teams I was in but it could be better. Almost all models created require the basic HTTP verbs and supported actions (`PUT`, `DELETE` etc) but they are still coded by the developers. No more, I say!

## HOW?

Simple, read in a `Swagger` file and a config file and OpenAPE will do all the `code monkey` stuff for you: building the routes, adding the models to a database, validation etc etc.

## HOW [CAN I HELP]?

PR the heck out of this. This is a super early version (work started end of July 2018) and I do not pretend to be any form of Go expert. 

If you are a beginner, check out the `TODO.md` file or look at open issues.

---

GET STARTED

* As of now, this is not a Go package and is still in early stages. It will be packaged soon.

1. `git clone https://github.com/encima/openape`

2. Replace or use the config file and the swagger file (`json` format)

3. `go run server.go`

