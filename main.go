package main

import (
	_ "github.com/joho/godotenv/autoload" // Autoload .env file.
	"log"
	"os"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/cmd/app"
)

// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token
//
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333" // fallback to default port
	}
	err := os.Setenv("API_PORT", port)
	if err != nil {
		log.Fatalf("Failed to set API_PORT environment variable: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("Application failed to start: %v", err)
	}
}
