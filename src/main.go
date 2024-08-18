package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

func runQuery(uri, username, password string) (_ []string, err error) {
	ctx := context.Background()
	config := func(conf *config.Config) {
		conf.MaxConnectionLifetime = 1 * time.Hour
		conf.MaxConnectionPoolSize = 50
		conf.ConnectionAcquisitionTimeout = 2 * time.Minute
		conf.SocketConnectTimeout = 2 * time.Minute
		conf.SocketKeepalive = true
	}
	fmt.Println("Connecting to Neo4j")
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""), config)
	if err != nil {
		fmt.Println("Failed to create driver:", err)
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}
	defer func() { err = handleClose(ctx, driver, err) }()

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	fmt.Println("Driver created successfully")

	query := `MATCH (m:Movie {title: $movieTitle})<-[:ACTED_IN]-(a:Actor) RETURN a.name AS actorName`
	params := map[string]any{"movieTitle": "The Matrix"}
	fmt.Println("Executing query:", query)
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		txResult, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		records, err := txResult.Collect(ctx)
		if err != nil {
			return nil, err
		}
		return records, nil
	})
	if err != nil {
		fmt.Println("Failed to execute query:", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	fmt.Println("Query executed successfully")

	records, ok := result.([]*neo4j.Record)
	if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	actorNames := make([]string, len(records))
	for i, record := range records {
		name, _, err := neo4j.GetRecordValue[string](record, "actorName")
		if err != nil {
			fmt.Println("Failed to get record value:", err)
			return nil, fmt.Errorf("failed to get record value: %w", err)
		}
		actorNames[i] = name
		fmt.Println("Retrieved actor name:", name)
	}
	return actorNames, nil
}

func handleClose(ctx context.Context, closer interface{ Close(context.Context) error }, previousError error) error {
	fmt.Println("Closing driver")
	err := closer.Close(ctx)
	if err == nil {
		fmt.Println("Driver closed successfully")
		return previousError
	}
	if previousError == nil {
		fmt.Println("Error closing driver:", err)
		return err
	}
	fmt.Println("Closure error occurred:", err, "Initial error was:", previousError)
	return fmt.Errorf("%v closure error occurred:\n%s\ninitial error was:\n%w", reflect.TypeOf(closer), err.Error(), previousError)
}

func main() {
	// results, err := runQuery("neo4j://localhost:7687", "neo4j", "testingpassword")
	// if err != nil {
	// 	fmt.Println("Failed to execute query:", err)
	// 	return
	// }
	// for _, result := range results {
	// 	fmt.Println(result)
	// }

	router()
}

func router() {
	f, _ := os.Create("../log/server.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})
	router.GET("/actor", func(c *gin.Context) {
		results, err := runQuery("neo4j://demo-neo4j:7687", "neo4j", "testingpassword")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, result := range results {
			c.JSON(http.StatusOK, gin.H{"result": result})
		}
	})

	// サーバーの起動状態を表示
	if err := router.Run("0.0.0.0:8080"); err != nil {
		fmt.Println("サーバーの起動に失敗しました:", err)
	} else {
		fmt.Println("サーバーが正常に起動しました")
	}
}