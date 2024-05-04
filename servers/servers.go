package servers

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
)

var (
	redisClient *redis.Client
)

// Data represents the structure of data loaded from the YAML file
type Data struct {
	Map     map[string]string `yaml:"map"`
	Servers []string          `yaml:"servers"`
}

// Init initializes a new Redis client, populates it with data from the YAML file, and returns the client
func Init() *redis.Client {
	// Connect to Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // Redis server password (leave empty if no password)
		DB:       0,                // Use the default DB
	})

	err := redisClient.FlushAll(context.Background()).Err()
	if err != nil {
		log.Fatalf("Failed to flush Redis: %v", err)
	}
	// Load data from YAML file
	data, err := loadDataFromYAML("./servers/servers.yaml")
	if err != nil {
		log.Fatal("Error loading data from YAML file:", err)
	}

	// Populate Redis with data from YAML file
	populateRedis(data)

	return redisClient
}

// loadDataFromYAML loads data from a YAML file into memory
func loadDataFromYAML(filename string) (*Data, error) {
	data := &Data{}
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, data)
	if err != nil {
		return nil, err
	}
	log.Println("Data loaded from YAML file:", data)

	return data, nil
}

// populateRedis populates the Redis cache with data from the YAML file
func populateRedis(data *Data) {
	// Populate map

	err := redisClient.HSet(context.Background(), "sessions", "dummy", "").Err()
	if err != nil {
		log.Fatal("Error initializing map in Redis:", err)
	}

	// Populate list
	err = redisClient.LPush(context.Background(), "servers", data.Servers).Err()
	if err != nil {
		log.Fatal("Error populating list in Redis:", err)
	}
}
