package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Neo4jConfiguration holds the configuration for connecting to the DB
type Neo4jConfiguration struct {
    URL      string
    Username string
    Password string
    Database string
}

// Movie represents a movie node in Neo4j
type Movie struct {
    Released int64  `json:"released"`
    Title    string `json:"title"`
}

// newDriver returns a connection to the DB
func (nc *Neo4jConfiguration) newDriver() (neo4j.Driver, error) {
    return neo4j.NewDriver(nc.URL, neo4j.BasicAuth(nc.Username, nc.Password, ""))
}

// getDataHandler will query the database and return the result as JSON
func getDataHandler(driver neo4j.Driver) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := driver.NewSession(neo4j.SessionConfig{
            AccessMode:   neo4j.AccessModeRead,
            DatabaseName: "neo4j",
        })
        defer unsafeClose(session)

        query := `MATCH (movie:Movie) RETURN movie.title as title, movie.released as released LIMIT $limit`
        result, err := session.Run(query, map[string]interface{}{"limit": 10})
        if err != nil {
            log.Println("Error querying Neo4j", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
            return
        }

        var movies []Movie
        for result.Next() {
            record := result.Record()
            released, found := record.Get("released")
            if !found {
                log.Println("Released not found in record")
                continue
            }
            title, found := record.Get("title")
            if !found {
                log.Println("Title not found in record")
                continue
            }
            movies = append(movies, Movie{
                Released: released.(int64),
                Title:    title.(string),
            })
        }

        if err = result.Err(); err != nil {
            log.Println("Error iterating over result", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over result"})
            return
        }

        c.JSON(http.StatusOK, movies)
    }
}

// addDummyDataHandler inserts dummy movie data into the database
func addDummyDataHandler(driver neo4j.Driver) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := driver.NewSession(neo4j.SessionConfig{
            AccessMode:   neo4j.AccessModeWrite,
            DatabaseName: "neo4j",
        })
        defer unsafeClose(session)

        dummyData := []Movie{
            {Title: "The Matrix", Released: 1999},
            {Title: "Inception", Released: 2010},
            {Title: "Interstellar", Released: 2014},
        }

        for _, movie := range dummyData {
            _, err := session.Run(
                `CREATE (movie:Movie {title: $title, released: $released})`,
                map[string]interface{}{
                    "title":    movie.Title,
                    "released": movie.Released,
                })
            if err != nil {
                log.Println("Error inserting data into Neo4j", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"status": "Dummy data inserted"})
    }
}

func main() {
    configuration := parseConfiguration()
    driver, err := configuration.newDriver()
    if err != nil {
        log.Fatal(err)
    }
    defer unsafeClose(driver)

    router := gin.Default()
    router.GET("/movies", getDataHandler(driver))
    router.GET("/movies/dummy", addDummyDataHandler(driver))

    router.Run(":8080")
}

func parseConfiguration() *Neo4jConfiguration {
    return &Neo4jConfiguration{
        URL:      "neo4j://neo4j:7687",
		Username: "neo4j",
        Password: "testingpassword",
    }
}

func unsafeClose(closeable io.Closer) {
    if err := closeable.Close(); err != nil {
        log.Fatal(fmt.Errorf("could not close resource: %w", err))
    }
}