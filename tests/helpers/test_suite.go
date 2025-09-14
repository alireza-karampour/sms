package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alireza-karampour/sms/internal/streams"
	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

// TestConfig holds the test configuration loaded from SmsGW.yaml
type TestConfig struct {
	Postgres struct {
		Address  string `mapstructure:"address"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"postgres"`
	NATS struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"nats"`
}

// TestSuite provides common setup and teardown for tests
type TestSuite struct {
	DB       *pgxpool.Pool
	NATSConn *nats.Base
	TestDB   string
	Cleanup  func()
	Config   *TestConfig
}

// LoadTestConfig loads the test configuration from SmsGW.yaml
func LoadTestConfig() *TestConfig {
	// Set up viper to read the config file
	viper.SetConfigName("SmsGW")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	viper.SetConfigType("yaml")

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		// Fallback to defaults if config file is not found
		viper.SetDefault("api.postgres.address", "127.0.0.1")
		viper.SetDefault("api.postgres.port", 5434)
		viper.SetDefault("api.postgres.username", "root")
		viper.SetDefault("api.postgres.password", "1234")
		viper.SetDefault("api.nats.address", "127.0.0.1:4223")
	}

	// Create config struct and unmarshal
	config := &TestConfig{}

	// Load API section (which is what we use for tests)
	config.Postgres.Address = viper.GetString("api.postgres.address")
	config.Postgres.Port = viper.GetInt("api.postgres.port")
	config.Postgres.Username = viper.GetString("api.postgres.username")
	config.Postgres.Password = viper.GetString("api.postgres.password")
	config.NATS.Address = viper.GetString("api.nats.address")

	return config
}

// SetupTestSuite initializes the test environment
func SetupTestSuite() *TestSuite {
	// Set test environment
	os.Setenv("GIN_MODE", "test")

	// Load test configuration
	config := LoadTestConfig()

	// Generate unique test database name
	testDB := fmt.Sprintf("sms_test_%d", time.Now().UnixNano())

	// Connect to PostgreSQL
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres",
		config.Postgres.Username,
		config.Postgres.Password,
		config.Postgres.Address,
		config.Postgres.Port,
	)

	pool, err := pgxpool.New(context.Background(), dbURL)
	Expect(err).NotTo(HaveOccurred())

	// Create test database
	_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", testDB))
	Expect(err).NotTo(HaveOccurred())

	// Close connection to default database
	pool.Close()

	// Connect to test database
	testDBURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		config.Postgres.Username,
		config.Postgres.Password,
		config.Postgres.Address,
		config.Postgres.Port,
		testDB,
	)

	testPool, err := pgxpool.New(context.Background(), testDBURL)
	Expect(err).NotTo(HaveOccurred())

	// Run schema migrations
	err = runSchemaMigrations(testPool)
	Expect(err).NotTo(HaveOccurred())

	// Connect to NATS
	natsConnRaw, err := nats.Connect(config.NATS.Address)
	Expect(err).NotTo(HaveOccurred())

	natsConn, err := nats.NewBase(natsConnRaw)
	Expect(err).NotTo(HaveOccurred())

	cleanup := func() {
		testPool.Close()
		natsConn.Close()

		// Drop test database
		cleanupPool, err := pgxpool.New(context.Background(), dbURL)
		if err == nil {
			cleanupPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s", testDB))
			cleanupPool.Close()
		}
	}

	return &TestSuite{
		DB:       testPool,
		NATSConn: natsConn,
		TestDB:   testDB,
		Cleanup:  cleanup,
		Config:   config,
	}
}

// SetupGinkgoSuite sets up Ginkgo test suite with common configuration
func SetupGinkgoSuite(t *testing.T, suiteName string) {
	RegisterFailHandler(ginkgo.Fail)

	// Set up viper configuration
	viper.SetConfigName("SmsGW")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	viper.SetConfigType("yaml")

	// Try to read config file, but don't fail if it doesn't exist
	err := viper.ReadInConfig()
	if err != nil {
		// If config file doesn't exist, set defaults
		viper.SetDefault("api.postgres.address", "postgres-e2e")
		viper.SetDefault("api.postgres.port", 5432)
		viper.SetDefault("api.postgres.username", "root")
		viper.SetDefault("api.postgres.password", "1234")
		viper.SetDefault("api.nats.address", "nats-e2e:4222")
	}

	ginkgo.RunSpecs(t, suiteName)
}

// runSchemaMigrations runs the database schema
func runSchemaMigrations(pool *pgxpool.Pool) error {
	schema := `
	CREATE SCHEMA IF NOT EXISTS public;

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		balance DECIMAL(10, 2) DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS phone_numbers (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users (id),
		phone_number VARCHAR(255) NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS sms (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users (id),
		phone_number_id INT NOT NULL REFERENCES phone_numbers (id),
		to_phone_number VARCHAR(255) NOT NULL,
		message VARCHAR(255) NOT NULL,
		status VARCHAR(255) NOT NULL DEFAULT 'pending',
		delivered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := pool.Exec(context.Background(), schema)
	return err
}

// CleanupTestData cleans up test data from database and NATS streams
func (ts *TestSuite) CleanupTestData() {
	ctx := context.Background()

	// Clean up database in reverse order of dependencies
	ts.DB.Exec(ctx, "DELETE FROM sms")
	ts.DB.Exec(ctx, "DELETE FROM phone_numbers")
	ts.DB.Exec(ctx, "DELETE FROM users")

	// Reset sequences
	ts.DB.Exec(ctx, "ALTER SEQUENCE users_id_seq RESTART WITH 1")
	ts.DB.Exec(ctx, "ALTER SEQUENCE phone_numbers_id_seq RESTART WITH 1")
	ts.DB.Exec(ctx, "ALTER SEQUENCE sms_id_seq RESTART WITH 1")

	// Clean up NATS streams
	ts.CleanupNATSStreams(ctx)
}

// CleanupNATSStreams removes all messages from NATS streams
func (ts *TestSuite) CleanupNATSStreams(ctx context.Context) {
	if ts.NATSConn == nil || ts.NATSConn.JetStream == nil {
		return
	}

	// List all streams and purge them
	streamNames := []string{
		streams.NORMAL_SMS_CONSUMER_NAME,
		streams.EXPRESS_SMS_CONSUMER_NAME,
	}

	for _, streamName := range streamNames {
		// Get the stream interface
		stream, err := ts.NATSConn.JetStream.Stream(ctx, streamName)
		if err != nil {
			// Stream might not exist, which is fine
			continue
		}

		// Purge all messages from the stream
		err = stream.Purge(ctx)
		if err != nil {
			// Log error but don't fail the test
			continue
		}
	}
}
