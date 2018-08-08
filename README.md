# OpenAPE
[![Build Status](https://travis-ci.org/encima/openape.svg?branch=master)](https://travis-ci.org/encima/openape)
[![Go Report Card](https://goreportcard.com/badge/github.com/encima/openape)](https://goreportcard.com/report/github.com/encima/openape)
[![GoDoc](https://godoc.org/github.com/encima/openape?status.svg)](https://godoc.org/github.com/encima/openape)

## WHAT?

OpenAPE is designed to be a server extension of the `OpenAPI` specification, reading in a `Swagger` file and creating routes with the specified security.

This already exists, you say? Why yes, yes it does. Many exist for OpenAPI 2 but few support OpenAPI3 and all of them require code to be written...

## WHY?

How much of web development consists of creating and maintaining an API that is marginally different from the last one? `Swagger` and `OpenAPI` has made this much easier with code generation tools and frameworks such as `Loopback`. Almost all models created require the basic HTTP verbs and supported actions (`PUT`, `DELETE` etc) but they are still coded by the developers. No more, We say!

## HOW?

Simple, read in a `Swagger` file and a config file and OpenAPE will do all the `code monkey` stuff for you: building the routes, adding the models to a database, validation etc etc.

## HOW [CAN I HELP]?

PR the heck out of this. This is a super early version (work started end of July 2018) and I do not pretend to be any form of Go expert. 

If you are a beginner, check out the `TODO.md` file or look at open issues.

---

## Requirements

* Postgres (remote or local)
* Go
* A config file (example can be found in the `config` folder

## GET STARTED

1. `go get github.com/encima/openape`
2. ```
    package main

    import (
        "github.com/encima/openape"
    )

    func main() {
        openape.NewServer("PATH/TO/CONFIG")
    }
    ```

