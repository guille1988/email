package integration

import (
	"context"
	"email/internal/bootstrap"
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/database"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

var TestApp *app.App
var TestHandler http.Handler
var TestConfig *config.Config
var MailpitApiPort int

// RunTests handles the integration tests setup, execution, and cleanup.
func RunTests(test *testing.M) {
	ctx := context.Background()

	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	TestConfig = setupConfig()
	mysqlInstance := setupDatabaseContainer(ctx, TestConfig)
	rabbitInstance := setupRabbitContainer(ctx, TestConfig)
	mailpitInstance := setupMailContainer(ctx, TestConfig)

	setupApplication(TestConfig)

	code := test.Run()

	_ = mysqlInstance.Terminate(ctx)
	_ = rabbitInstance.Terminate(ctx)
	_ = mailpitInstance.Terminate(ctx)
	os.Exit(code)
}

// setupConfig initializes and returns the application configuration.
func setupConfig() *config.Config {
	cfg, _ := config.New()
	cfg.App.Env = config.TestingEnv
	return cfg
}

// setupDatabaseContainer starts a MySQL container and updates the configuration.
func setupDatabaseContainer(ctx context.Context, cfg *config.Config) *mysql.MySQLContainer {
	databaseConfig := cfg.Database.Connections[config.Default]
	testDatabaseName := databaseConfig.Database + "_test"

	mysqlInstance, err := mysql.Run(ctx, "mysql:8.0",
		mysql.WithDatabase(testDatabaseName),
		mysql.WithUsername(databaseConfig.Username),
		mysql.WithPassword(databaseConfig.Password),
	)

	if err != nil {
		panic(err)
	}

	host, _ := mysqlInstance.Host(ctx)
	port, _ := mysqlInstance.MappedPort(ctx, "3306")

	databaseConfig.Host = host
	databaseConfig.Port = port.Port()
	databaseConfig.Database = testDatabaseName
	cfg.Database.Connections[config.Default] = databaseConfig

	return mysqlInstance
}

// setupRabbitContainer starts a RabbitMQ container and updates the configuration.
func setupRabbitContainer(ctx context.Context, cfg *config.Config) *rabbitmq.RabbitMQContainer {
	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:3-management-alpine",
		rabbitmq.WithAdminPassword(cfg.RabbitMQ.Password),
		rabbitmq.WithAdminUsername(cfg.RabbitMQ.User),
	)

	if err != nil {
		panic(err)
	}

	host, _ := rabbitContainer.Host(ctx)
	port, _ := rabbitContainer.MappedPort(ctx, nat.Port(cfg.RabbitMQ.Port))

	cfg.RabbitMQ.Host = host
	cfg.RabbitMQ.Port = port.Port()

	return rabbitContainer
}

// setupMailContainer starts a Mailpit container and updates the configuration.
func setupMailContainer(ctx context.Context, cfg *config.Config) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForHTTP("/api/v1/info").WithPort("8025/tcp"),
	}
	mailpitContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		panic(err)
	}

	host, _ := mailpitContainer.Host(ctx)
	smtpPort, _ := mailpitContainer.MappedPort(ctx, "1025")
	apiPort, _ := mailpitContainer.MappedPort(ctx, "8025")

	cfg.Mail.Host = host
	cfg.Mail.Port = smtpPort.Int()
	MailpitApiPort = apiPort.Int()

	return mailpitContainer
}

// setupApplication initializes the app instance and runs database migrations.
func setupApplication(cfg *config.Config) {
	appInstance, err := bootstrap.NewTestingApp(cfg)

	if err != nil {
		panic(err)
	}

	var migration *migrate.Migrate
	migration, err = database.NewMigration(*cfg, config.Default)

	if err != nil {
		panic(err)
	}

	err = database.Migrate(migration)

	if err != nil {
		panic(err)
	}

	TestApp = appInstance
	TestHandler = bootstrap.NewTestingHandler(appInstance)
}

// RefreshDatabase resets the database to a clean state.
func RefreshDatabase() {
	migration, err := database.NewMigration(*TestConfig, config.Default)

	if err != nil {
		panic(err)
	}

	err = migration.Down()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

	err = database.Migrate(migration)

	if err != nil {
		panic(err)
	}
}

// ExecuteRequest performs an HTTP request against the global test handler and returns the response recorder.
func ExecuteRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	TestHandler.ServeHTTP(recorder, request)

	return recorder
}

// TestCase is a wrapper that runs RefreshDatabase before executing the actual test logic.
func TestCase(test *testing.T, name string, testFunction func(test *testing.T)) {
	test.Run(name, func(test *testing.T) {
		RefreshDatabase()
		testFunction(test)
	})
}
