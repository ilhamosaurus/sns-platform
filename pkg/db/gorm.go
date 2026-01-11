package db

import (
	"fmt"
	"log"
	"time"

	"github.com/ilhamosaurus/sns-platform/internal/model"
	"github.com/ilhamosaurus/sns-platform/pkg/types"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
)

// Config holds database configuration for all supported databases
type Config struct {
	Type     DatabaseType `yaml:"type"` // postgres, mysql, sqlite
	Host     string       `yaml:"host"`
	Port     string       `yaml:"port"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	DBName   string       `yaml:"dbname"`
	SSLMode  string       `yaml:"sslmode"`  // For PostgreSQL
	Charset  string       `yaml:"charset"`  // For MySQL
	FilePath string       `yaml:"filepath"` // For SQLite

	// Connection pool settings
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`

	// GORM settings
	LogLevel       string `yaml:"log_level"` // silent, error, warn, info
	PrepareStmt    bool   `yaml:"prepare_stmt"`
	SkipDefaultTxn bool   `yaml:"skip_default_txn"`
}

var db *gorm.DB

// Initialize establishes database connection with optimized settings
func Initialize(config Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	var err error

	// Select appropriate database driver based on type
	switch config.Type {
	case PostgreSQL:
		dialector, err = getPostgresDialector(config)
	case MySQL:
		dialector, err = getMySQLDialector(config)
	case SQLite:
		dialector, err = getSQLiteDialector(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create database dialector: %w", err)
	}

	// Set GORM logger level
	logLevel := getLogLevel(config.LogLevel)

	// Open database connection
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		TranslateError:         true,
		PrepareStmt:            config.PrepareStmt,
		SkipDefaultTransaction: config.SkipDefaultTxn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Apply connection pool settings
	maxIdleConns := config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	maxOpenConns := config.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 100
	}
	connMaxLifetime := config.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour
	}
	connMaxIdleTime := config.ConnMaxIdleTime
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 10 * time.Minute
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	log.Printf("✓ Database connection established successfully (Type: %s)", config.Type)
	return db, nil
}

// getPostgresDialector creates PostgreSQL dialector
func getPostgresDialector(config Config) (gorm.Dialector, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		getSSLMode(config.SSLMode),
	)

	log.Printf("Connecting to PostgreSQL: %s:%s/%s", config.Host, config.Port, config.DBName)
	return postgres.Open(dsn), nil
}

// getMySQLDialector creates MySQL dialector
func getMySQLDialector(config Config) (gorm.Dialector, error) {
	charset := config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		charset,
	)

	log.Printf("Connecting to MySQL: %s:%s/%s", config.Host, config.Port, config.DBName)
	return mysql.Open(dsn), nil
}

// getSQLiteDialector creates SQLite dialector
func getSQLiteDialector(config Config) (gorm.Dialector, error) {
	filePath := config.FilePath
	if filePath == "" {
		filePath = "social_media.db"
	}

	log.Printf("Connecting to SQLite: %s", filePath)
	return sqlite.Open(filePath), nil
}

// getSSLMode returns appropriate SSL mode or default
func getSSLMode(sslMode string) string {
	if sslMode == "" {
		return "disable"
	}
	return sslMode
}

// getLogLevel converts string log level to GORM logger level
func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// Migrate runs all database migrations
func Migrate() error {
	log.Println("Running database migrations...")

	// Auto-migrate all model
	err := db.AutoMigrate(
		&model.User{},
		&model.Follow{},
		&model.Post{},
		&model.Comment{},
		&model.Reaction{},
		&model.Message{},
		&model.Notification{},
		&model.ActivityFeed{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get database type
	dbType := getDatabaseType()

	// Create database-specific additional indexes
	if err := createAdditionalIndexes(dbType); err != nil {
		log.Printf("Warning: Failed to create some additional indexes: %v", err)
		// Don't return error - some indexes might not be supported
	}

	// Create composite indexes
	if err := createCompositeIndexes(dbType); err != nil {
		log.Printf("Warning: Failed to create some composite indexes: %v", err)
	}

	log.Println("✓ Database migrations completed successfully")
	return nil
}

// getDatabaseType returns the current database type
func getDatabaseType() DatabaseType {
	dbName := db.Name()
	switch dbName {
	case "postgres":
		return PostgreSQL
	case "mysql":
		return MySQL
	case "sqlite":
		return SQLite
	default:
		return SQLite
	}
}

// createAdditionalIndexes creates performance-critical indexes
func createAdditionalIndexes(dbType DatabaseType) error {
	switch dbType {
	case PostgreSQL:
		return createPostgresIndexes()
	case MySQL:
		return createMySQLIndexes()
	case SQLite:
		return createSQLiteIndexes()
	default:
		return nil
	}
}

// createPostgresIndexes creates PostgreSQL-specific indexes
func createPostgresIndexes() error {
	log.Println("Creating PostgreSQL-specific indexes...")

	// Enable pg_trgm extension for fuzzy search
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		log.Printf("Warning: Could not create pg_trgm extension: %v", err)
	}

	// Trigram index for username search
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin(username gin_trgm_ops)").Error; err != nil {
		log.Printf("Warning: Could not create trigram index on username: %v", err)
	}

	// Index for post feed queries (most recent posts)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_created_desc ON posts (created_at DESC) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}

	// Index for notification queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, is_read, created_at DESC) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}

	// Index for message conversations
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages (sender_id, receiver_id, created_at DESC) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}

	// Index for unread messages count
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages (receiver_id, is_read) WHERE deleted_at IS NULL AND is_read = false").Error; err != nil {
		return err
	}

	// Partial index for public posts
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_public ON posts (created_at DESC) WHERE is_public = true AND deleted_at IS NULL").Error; err != nil {
		return err
	}

	log.Println("✓ PostgreSQL-specific indexes created")
	return nil
}

// createMySQLIndexes creates MySQL-specific indexes
func createMySQLIndexes() error {
	log.Println("Creating MySQL-specific indexes...")

	// MySQL doesn't support partial indexes, so we create regular indexes

	// Index for post feed queries
	if err := db.Exec("CREATE INDEX idx_posts_created_desc ON posts (created_at DESC)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	// Composite index for notifications
	if err := db.Exec("CREATE INDEX idx_notifications_user_unread ON notifications (user_id, is_read, created_at)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	// Index for message conversations
	if err := db.Exec("CREATE INDEX idx_messages_conversation ON messages (sender_id, receiver_id, created_at)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	// Full-text index for username search (MySQL alternative to pg_trgm)
	if err := db.Exec("CREATE FULLTEXT INDEX idx_users_username_fulltext ON users (username, full_name)").Error; err != nil {
		log.Printf("Warning: Could not create fulltext index: %v", err)
	}

	log.Println("✓ MySQL-specific indexes created")
	return nil
}

// createSQLiteIndexes creates SQLite-specific indexes
func createSQLiteIndexes() error {
	log.Println("Creating SQLite-specific indexes...")

	// SQLite has limited index features, create basic indexes

	// Index for post feed queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_created_desc ON posts (created_at DESC)").Error; err != nil {
		return err
	}

	// Index for notifications
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, is_read, created_at)").Error; err != nil {
		return err
	}

	// Index for messages
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages (sender_id, receiver_id, created_at)").Error; err != nil {
		return err
	}

	log.Println("✓ SQLite-specific indexes created")
	return nil
}

// createCompositeIndexes creates composite indexes for complex queries
func createCompositeIndexes(dbType DatabaseType) error {
	log.Println("Creating composite indexes...")

	switch dbType {
	case PostgreSQL:
		return createPostgresCompositeIndexes()
	case MySQL:
		return createMySQLCompositeIndexes()
	case SQLite:
		return createSQLiteCompositeIndexes()
	}
	return nil
}

// createPostgresCompositeIndexes creates PostgreSQL composite indexes
func createPostgresCompositeIndexes() error {
	// Composite index for activity feed ordering
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_feed_user_time ON activity_feeds (user_id, post_created DESC) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}

	// Composite index for reaction counts
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_target_type ON reactions (post_id, type) WHERE post_id IS NOT NULL AND deleted_at IS NULL").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_comment_type ON reactions (comment_id, type) WHERE comment_id IS NOT NULL AND deleted_at IS NULL").Error; err != nil {
		return err
	}

	return nil
}

// createMySQLCompositeIndexes creates MySQL composite indexes
func createMySQLCompositeIndexes() error {
	// MySQL composite indexes without partial conditions
	if err := db.Exec("CREATE INDEX idx_activity_feed_user_time ON activity_feeds (user_id, post_created)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	if err := db.Exec("CREATE INDEX idx_reactions_post_type ON reactions (post_id, type)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	if err := db.Exec("CREATE INDEX idx_reactions_comment_type ON reactions (comment_id, type)").Error; err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	return nil
}

// createSQLiteCompositeIndexes creates SQLite composite indexes
func createSQLiteCompositeIndexes() error {
	// SQLite composite indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_feed_user_time ON activity_feeds (user_id, post_created)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_post_type ON reactions (post_id, type)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_comment_type ON reactions (comment_id, type)").Error; err != nil {
		return err
	}

	return nil
}

// Seed populates database with sample data for testing
func Seed() error {
	log.Println("Seeding database with sample data...")

	// Check if data already exists
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data, skipping seed")
		return nil
	}

	// Create sample users
	users := []model.User{
		{
			Username:     "alice_wonder",
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$EXAMPLE_HASH_1",
			FullName:     "Alice Wonderland",
			Bio:          "Tech enthusiast and coffee lover ☕",
			IsVerified:   true,
		},
		{
			Username:     "bob_builder",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$EXAMPLE_HASH_2",
			FullName:     "Bob Builder",
			Bio:          "Building the future, one line at a time",
			IsVerified:   false,
		},
		{
			Username:     "charlie_dev",
			Email:        "charlie@example.com",
			PasswordHash: "$2a$10$EXAMPLE_HASH_3",
			FullName:     "Charlie Developer",
			Bio:          "Full-stack developer | Open source contributor",
			IsVerified:   true,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Create follow relationships
	follows := []model.Follow{
		{FollowerID: users[0].ID, FollowingID: users[1].ID},
		{FollowerID: users[0].ID, FollowingID: users[2].ID},
		{FollowerID: users[1].ID, FollowingID: users[0].ID},
		{FollowerID: users[2].ID, FollowingID: users[0].ID},
	}

	if err := db.Create(&follows).Error; err != nil {
		return fmt.Errorf("failed to seed follows: %w", err)
	}

	// Create sample posts
	posts := []model.Post{
		{
			UserID:    users[0].ID,
			Content:   "Just finished an amazing project using Go and GORM! 🚀",
			MediaType: types.MediaTypeText,
			IsPublic:  true,
		},
		{
			UserID:    users[1].ID,
			Content:   "Check out this cool architecture diagram!",
			MediaType: types.MediaTypeImage,
			MediaURL:  "https://example.com/image1.jpg",
			IsPublic:  true,
		},
		{
			UserID:    users[2].ID,
			Content:   "Working on database optimization. Tips anyone?",
			MediaType: types.MediaTypeText,
			IsPublic:  true,
		},
	}

	if err := db.Create(&posts).Error; err != nil {
		return fmt.Errorf("failed to seed posts: %w", err)
	}

	// Create sample comments
	comments := []model.Comment{
		{
			PostID:  posts[0].ID,
			UserID:  users[1].ID,
			Content: "Great work! Would love to see the code.",
		},
		{
			PostID:  posts[0].ID,
			UserID:  users[2].ID,
			Content: "This is inspiring! Keep it up!",
		},
	}

	if err := db.Create(&comments).Error; err != nil {
		return fmt.Errorf("failed to seed comments: %w", err)
	}

	// Create sample reactions
	reactions := []model.Reaction{
		{UserID: users[1].ID, PostID: &posts[0].ID, Type: types.ReactionTypeLike},
		{UserID: users[2].ID, PostID: &posts[0].ID, Type: types.ReactionTypeLove},
		{UserID: users[0].ID, PostID: &posts[1].ID, Type: types.ReactionTypeLike},
	}

	if err := db.Create(&reactions).Error; err != nil {
		return fmt.Errorf("failed to seed reactions: %w", err)
	}

	log.Println("✓ Database seeded successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDatabaseInfo returns information about the current database connection
func GetDatabaseInfo() map[string]interface{} {
	sqlDB, _ := db.DB()
	stats := sqlDB.Stats()

	return map[string]interface{}{
		"type":                db.Name(),
		"max_open_conns":      stats.MaxOpenConnections,
		"open_conns":          stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}
