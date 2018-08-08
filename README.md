# OpenAPE
[![Build Status](https://travis-ci.com/encima/openape.svg?branch=master)](https://travis-ci.com/encima/openape)
[![Go Report Card](https://goreportcard.com/badge/github.com/encima/openape)](https://goreportcard.com/report/github.com/encima/openape)
[![GoDoc](https://godoc.org/github.com/encima/openape?status.svg)](https://godoc.org/github.com/encima/openape)

OpenAPE is designed to be a server extension of the `OpenAPI` specification, reading in a `Swagger` file and a config and OpenAPE will do all the `code monkey` stuff for you: building the routes, adding the models to a database, validation etc.

Much of web development consists of creating and maintaining an API that is marginally different from the last one? `Swagger` and `OpenAPI` has made this much easier with code generation tools and frameworks such as `Loopback`. Almost all models created require the basic HTTP verbs and supported actions (`PUT`, `DELETE` etc) but most existing tools only generate method stubs still requiring addition code from developers this aims to solve that.

---

## Contributing

PRs are welcome, this is a super early version and is far from perfect.

If you are a beginner, check out the [labelled open issues](https://github.com/encima/openape/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

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

