<h1 align="center">
  Go API Boilerplates - gin/gorm/auth-jwt
</h1>

<div align="center">

[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
![Tests](https://github.com/yizeng/gab/actions/workflows/test.yml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/yizeng/gab/gin/gorm/auth-jwt)](https://goreportcard.com/report/github.com/yizeng/gab/gin/gorm/auth-jwt)

</div>

<hr />

This is a basic example of Go API
with [gin-gonic/gin][gin-gonic/gin], PostgreSQL and [GORM][go-gorm/gorm].

A full list of libraries used can be found in [Dependencies](#dependencies) section.

## Table of Contents

- [Development](#development)
  + [Tools](#tools)
  + [Documentation](#documentation)
- [Dependencies](#dependencies)
  + [API](#api)
  + [Database](#database)
  + [Testing](#testing)

## Development

This project can be booted up either

- natively with Go installed locally (assuming dependencies like PostgresSQL, Redis, etc. also exist)
- or by Docker Compose (PostgreSQL, Redis, etc. are included out of the box)

Please refer to [Makefile](./Makefile) or [docker-compose.yml](./docker-compose.yml) for details.

By default, the server will run at <http://localhost:3333>, but everything is configurable via [.env](.env) file.

### Tools

Some common development tools are introduced for better local development experience.

- [cosmtrek/air](https://github.com/cosmtrek/air) - Live reload for Go apps
- [GoTestTools/gotestfmt](https://github.com/GoTestTools/gotestfmt) - go test output for humans
- [swaggo/swag][swaggo/swag] - Automatically generate RESTful API documentation with Swagger 2.0 for Go.

```
go install github.com/cosmtrek/air@latest
go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

Alternatively, use `make install` to install the required Go tooling locally.

### Documentation

This service has integrated [swaggo/swag][swaggo/swag] to create OpenAPI documentation automatically.

Use command `make generate/api` or `swag init` to run the generation.

Then navigate to <http://localhost:3333/swagger/index.html> to view the API documentation.

## Dependencies

### API

- [gin-gonic/gin][gin-gonic/gin] - Gin is a HTTP web framework written in Go (Golang).
  - [gin-contrib/cors][gin-contrib/cors] - Official CORS gin's middleware
  - [gin-contrib/requestid][gin-contrib/requestid] - Request ID middleware for Gin Framework
- [joho/godotenv][joho/godotenv] - A Go port of Ruby's dotenv library (Loads environment variables from .env files)
- [spf13/viper][spf13/viper] - Go configuration with fangs
- [uber-go/zap][uber-go/zap] - Blazing fast, structured, leveled logging in Go.
- [go-ozzo/ozzo-validation][go-ozzo/ozzo-validation] - An idiomatic Go (golang) validation package. Supports configurable and extensible validation rules (validators) using normal language constructs instead of error-prone struct tags.
- [swaggo/gin-swagger][swaggo/gin-swagger] - gin middleware to automatically generate RESTful API documentation with Swagger 2.0.
- [dlclark/regexp2][dlclark/regexp2] - A full-featured regex engine in pure Go based on the .NET engine
- [golang-jwt/jwt][golang-jwt/jwt] - Golang implementation of JSON Web Tokens (JWT).

### Database
- [PostgreSQL][PostgreSQL] - The World's Most Advanced Open Source Relational Database
- [go-gorm/gorm][go-gorm/gorm] - The fantastic ORM library for Golang, aims to be developer friendly

### Testing
- [stretchr/testify][stretchr/testify] - A toolkit with common assertions and mocks that plays nicely with the standard library
- [ory/dockertest][ory/dockertest] - Write better integration tests! Dockertest helps you boot up ephermal docker images for your Go tests with minimal work.

[gin-gonic/gin]: https://github.com/gin-gonic/gin
[gin-contrib/cors]: https://github.com/gin-contrib/cors
[gin-contrib/requestid]: https://github.com/gin-contrib/requestid
[joho/godotenv]: https://github.com/joho/godotenv
[spf13/viper]: https://github.com/spf13/viper
[uber-go/zap]: https://github.com/uber-go/zap
[stretchr/testify]: https://github.com/stretchr/testify
[go-ozzo/ozzo-validation]: https://github.com/go-ozzo/ozzo-validation
[PostgreSQL]: https://www.postgresql.org/
[go-gorm/gorm]: https://github.com/go-gorm/gorm
[ory/dockertest]: https://github.com/ory/dockertest
[swaggo/swag]: https://github.com/swaggo/swag
[swaggo/gin-swagger]: https://github.com/swaggo/gin-swagger
[dlclark/regexp2]: https://github.com/dlclark/regexp2
[golang-jwt/jwt]: https://github.com/golang-jwt/jwt
