package docs

//go:generate oapi-codegen -generate "types,strict-server,embedded-spec,gin" -package handlers -o ../internal/handlers/api.gen.go openapi.yml
