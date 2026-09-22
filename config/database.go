package database

import (
	"fmt"
	"os"
	"pos-master/migrations"
	"pos-master/models"
	"pos-master/seeders"

	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func GetDBConfig() DBConfig {

	return DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		Name:     os.Getenv("DB_NAME"),
	}
}

func GetDSN(cfg DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}
func InitDB() {
	var err error

	// dsn := "root@tcp(127.0.0.1:3306)/bus_ticketing_system?charset=utf8mb4&parseTime=True&loc=Local"
	// dsn := "sandbox_user:sandbox_password@tcp(10.139.40.25:3306)/pgsandbox?charset=utf8mb4&parseTime=True&loc=Local"

	// cfg := GetDBConfig()
	// dsn := GetDSN(cfg)
	dsn := os.Getenv("DB_URL")

	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	ResetDatabaseIfNeeded(DB)

	err = DB.AutoMigrate(

		&models.Role{},       // Migrate `Role` first as `User` depends on it
		&models.Permission{}, // Other independent tables can be migrated here
		&models.TerminalType{},
		&models.App{},
		&models.AppVersion{},
		&models.PosDevice{},
		&models.LocationHistory{},
		&models.Event{},
		&models.Business{},
	)

	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// users is a shared table. AutoMigrate it only when it does not exist so
	// a fresh install still boots; on the live database we never mutate its
	// schema except through versioned migrations.
	if !DB.Migrator().HasTable(&models.User{}) {
		if err := DB.AutoMigrate(&models.User{}); err != nil {
			log.Fatalf("failed to create users table: %v", err)
		}
	}

	if err := migrations.Run(DB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := seeders.SeedRoles(DB); err != nil {
		log.Fatalf("failed to seed roles: %v", err)
	}

	if err := seeders.SeedAdminUser(DB); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	if err := seeders.SeedTerminalTypes(DB); err != nil {
		log.Fatalf("failed to seed terminal types: %v", err)
	}

	fmt.Println("Database connected successfully")
}

func LoadEnv() {

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, relying on system environment variables")
	}

}

func ResetDatabaseIfNeeded(db *gorm.DB) {
	reset := os.Getenv("RESET_DB")

	if reset != "true" && reset != "1" {
		return
	}

	log.Println("⚠️ RESET_DB is true — dropping all tables...")

	err := db.Migrator().DropTable(
		// Auth & RBAC
		&models.Permission{},
		&models.Role{},
		&models.User{},
		// App & Device Management
		&models.App{},
		&models.AppVersion{},
		&models.TerminalType{},
		&models.PosDevice{},
		&models.LocationHistory{},
		// Event & Business Management
		&models.Event{},
		&models.Business{},
	)

	if err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}

	log.Println("✅ All tables dropped successfully")
}
