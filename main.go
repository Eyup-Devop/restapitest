package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

func init() {
	flag.Parse()
}

func main() {
	connectionStr := GetEnv("CONNECTION", "example connection string not found")
	testSecret := GetEnv("TESTSECRET", "example secret not found")
	log.Println("Connection:", connectionStr)
	log.Println("Secret:", testSecret)
	server := CreateServer()
	Start(server)
}

func CreateServer() *fiber.App {
	mainRoot := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		ProxyHeader:             "X-Real-IP",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	})

	mainRoot.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))
	mainRoot.Use(RegisterAccessLogs())

	mainRoot.Get("/testapi", TestApiHandler)
	return mainRoot
}

func RegisterAccessLogs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		returnErr := c.Next()
		var clientIp string = ""
		if len(c.IPs()) == 0 {
			clientIp = c.IP()
		} else {
			clientIp = c.IPs()[0]
		}

		log.Printf("%s - %s - %s \n", clientIp, c.Method(), c.Path())
		return returnErr
	}
}

func Start(app *fiber.App) {
	err := app.Listen(":9090")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func TestApiHandler(c *fiber.Ctx) error {
	response := &struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}{
		FirstName: "John",
		LastName:  "Smith",
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"rest_api_test": response}})
}

func GetEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
