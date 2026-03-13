package testutils

import (
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB replaces the global DB with an in-memory SQLite database.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	database.DB = db
	return db
}

