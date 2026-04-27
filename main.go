package main

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/alivehamster/transcodarr/libs"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	port := "8080"

	dir, err := filepath.Abs("./config")
	if err != nil {
		log.Fatal("Failed to resolve current directory:", err)
	}

	db, err := sql.Open("sqlite3", filepath.Join(dir, "database.db"))
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS libraries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			cron TEXT,
			config TEXT
		);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	js := libs.NewJobScheduler()

	err = js.StartJobs(db)
	if err != nil {
		log.Fatal("Failed to start jobs:", err)
	}

	app := fiber.New()

	app.Get("/api/libraries", func(c fiber.Ctx) error {
		var libraries []libs.Library

		rows, err := db.Query("SELECT id, name, cron FROM libraries")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch libraries"})
		}
		defer rows.Close()

		for rows.Next() {
			var lib libs.Library
			if err := rows.Scan(&lib.ID, &lib.Name, &lib.Cron); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse library data"})
			}
			libraries = append(libraries, lib)
		}
		if err := rows.Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error iterating library data"})
		}

		return c.JSON(libraries)
	})

	app.Get("/api/handbrakeProfiles", func(c fiber.Ctx) error {
		profiles, err := libs.GetHandBrakeProfiles()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch HandBrake profiles"})
		}
		return c.JSON(profiles)
	})

	app.Post("/api/createLibrary", func(c fiber.Ctx) error {
		var lib libs.Library
		if err := c.Bind().JSON(&lib); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		result, err := db.Exec("INSERT INTO libraries (name, cron, config) VALUES (?, ?, ?)", lib.Name, lib.Cron, lib.Config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create library"})
		}

		id, _ := result.LastInsertId()
		lib.ID = int(id)

		err = js.EditSchedule(db, lib)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to schedule job"})
		}

		return c.JSON(lib)
	})

	app.Post("/api/deleteLibrary", func(c fiber.Ctx) error {
		var lib libs.Library
		if err := c.Bind().JSON(&lib); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		_, err := db.Exec("DELETE FROM libraries WHERE id = ?", lib.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete library"})
		}

		js.DeleteJob(lib.ID)

		return c.SendStatus(fiber.StatusOK)

	})

	app.Get("/*", static.New("./frontend/dist"))

	log.Println("Server starting on", port)
	log.Fatal(app.Listen(":" + port))

}
