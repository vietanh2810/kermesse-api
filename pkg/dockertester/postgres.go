package dockertester

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	minPort          = 10000
	maxPort          = 65535
	postgresDBName   = "postgres"
	postgresUsername = "postgres"
	postgresPassword = "postgres"
	postgresPort     = "5432"
	postgresImage    = "postgres"
	postgresImageTag = "16.1-alpine"
)

type Dockertester struct {
	HostPort string
	Pool     *dockertest.Pool
	Resource *dockertest.Resource
}

// InitPostgres function initialize dockertest for PostgreSQL.
// Please refer to the official example here: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
func InitPostgres() *Dockertester {
	// create a random port number between min and max to avoid conflicts.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hostPort := fmt.Sprint(r.Intn(maxPort-minPort) + minPort)

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: postgresImage,
		Tag:        postgresImageTag,
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", postgresPassword),
			fmt.Sprintf("POSTGRES_USERNAME=%s", postgresUsername),
			fmt.Sprintf("POSTGRES_DB=%s", postgresDBName),
			"listen_addresses = '*'",
		},
		ExposedPorts: []string{hostPort},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {
				{HostIP: "0.0.0.0", HostPort: hostPort},
			},
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		// keep retrying if opening database fails.
		db, err := OpenPostgres(resource, hostPort)
		if err != nil {
			return err
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		return sqlDB.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return &Dockertester{
		HostPort: hostPort,
		Pool:     pool,
		Resource: resource,
	}
}

func OpenPostgres(resource *dockertest.Resource, port string) (*gorm.DB, error) {
	host := resource.GetBoundIP(fmt.Sprintf("%s/tcp", postgresPort))
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s sslmode=disable password=%s dbname=%s",
		host, port, postgresUsername, postgresPassword, postgresDBName,
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return gdb, err
}
