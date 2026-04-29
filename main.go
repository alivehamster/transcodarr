package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"path/filepath"
	"strconv"

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
			config TEXT,
			skiplist TEXT
		);
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT
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

	app.Get("/api/library/:id", func(c fiber.Ctx) error {
		idstr := c.Params("id")

		id, err := strconv.Atoi(idstr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid library ID"})
		}

		row := db.QueryRow("SELECT id, name, cron, config FROM libraries WHERE id = ?", id)
		var lib libs.Library
		var configJSON string
		if err := row.Scan(&lib.ID, &lib.Name, &lib.Cron, &configJSON); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Library not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch library"})
		}

		if err := json.Unmarshal([]byte(configJSON), &lib.Config); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse library config"})
		}

		return c.JSON(lib)
	})

	app.Get("/api/handbrakeProfiles", func(c fiber.Ctx) error {
		profiles, err := libs.GetHandBrakeProfiles()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch HandBrake profiles"})
		}
		return c.JSON(profiles)
	})

	app.Get("/api/skiplist/:id", func(c fiber.Ctx) error {
		idstr := c.Params("id")

		id, err := strconv.Atoi(idstr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid library ID"})
		}

		row := db.QueryRow("SELECT skiplist FROM libraries WHERE id = ?", id)
		var skiplistJSON sql.NullString
		var skiplist []libs.Skip
		if err := row.Scan(&skiplistJSON); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Library not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch library"})
		}

		if skiplistJSON.Valid && skiplistJSON.String != "" {
			if err := json.Unmarshal([]byte(skiplistJSON.String), &skiplist); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse library skiplist"})
			}
		}
		if skiplist == nil {
			skiplist = []libs.Skip{}
		}

		return c.JSON(skiplist)
	})

	app.Put("/api/editSkiplist", func(c fiber.Ctx) error {
		var skiplist libs.SkipList
		if err := c.Bind().JSON(&skiplist); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		skiplistJSON, err := json.Marshal(skiplist.Skips)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize skiplist"})
		}

		result, err := db.Exec("UPDATE libraries SET skiplist = ? WHERE id = ?", string(skiplistJSON), skiplist.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update library"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Library not found"})
		}

		return c.JSON(skiplist)
	})

	app.Post("/api/createLibrary", func(c fiber.Ctx) error {
		var lib libs.Library
		if err := c.Bind().JSON(&lib); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		configJSON, err := json.Marshal(lib.Config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize config"})
		}

		result, err := db.Exec("INSERT INTO libraries (name, cron, config) VALUES (?, ?, ?)", lib.Name, lib.Cron, string(configJSON))
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

	app.Put("/api/editLibrary", func(c fiber.Ctx) error {
		var lib libs.Library
		if err := c.Bind().JSON(&lib); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		configJSON, err := json.Marshal(lib.Config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize config"})
		}

		result, err := db.Exec("UPDATE libraries SET name = ?, cron = ?, config = ? WHERE id = ?", lib.Name, lib.Cron, string(configJSON), lib.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update library"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Library not found"})
		}

		err = js.EditSchedule(db, lib)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to schedule job"})
		}

		return c.JSON(lib)
	})

	app.Delete("/api/deleteLibrary/:id", func(c fiber.Ctx) error {
		idstr := c.Params("id")

		id, err := strconv.Atoi(idstr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid library ID"})
		}

		_, err = db.Exec("DELETE FROM libraries WHERE id = ?", id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete library"})
		}

		js.DeleteJob(id)

		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/api/run/:id", func(c fiber.Ctx) error {
		idstr := c.Params("id")

		id, err := strconv.Atoi(idstr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid library ID"})
		}

		go libs.RunJob(db, js, id)

		return c.JSON(fiber.Map{"message": "Job triggered"})
	})

	app.Get("/*", static.New("./frontend/dist"))

	log.Println("Server starting on", port)
	log.Fatal(app.Listen(":" + port))

}
