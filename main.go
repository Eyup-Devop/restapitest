package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func init() {
	flag.Parse()
}

var DBConn *sqlx.DB

func main() {
	connectionStr := GetEnv("CONNECTION_STRING", "failed to get connection string")
	testSecret := GetEnv("TESTSECRET", "example secret not found")
	log.Println("Secret:", testSecret)

	DBConn = ConnectDB(connectionStr)

	defer DBConn.Close()
	server := CreateServer()
	Start(server)
}

func ConnectDB(connStr string) *sqlx.DB {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil
	}

	return db
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
	api := mainRoot.Group("/api")

	api.Get("/testapi", TestApiHandler)
	api.Get("/connection", TestConnection)
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

func TestConnection(c *fiber.Ctx) error {
	if DBConn == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "failed", "data": "failed to connect DB"})
	}
	err := DBConn.Ping()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "failed", "data": fiber.Map{"connection_test": err.Error()}})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "success", "data": "connected"})
}

func GetEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
