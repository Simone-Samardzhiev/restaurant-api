package rest_test

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"restaurant/internal/test"
	"testing"

	"github.com/gin-gonic/gin"
)

// database connection used for integration tests.
// Only connected if the test flag short is not set.
var database *sql.DB

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	flag.Parse()
	if !testing.Short() {
		url, ok := os.LookupEnv("TEST_DB_URL")
		if !ok {
			log.Fatal("TEST_DB_URL not set")
		}

		database = test.ConnectToDb(url)
	}

	os.Exit(m.Run())
}
