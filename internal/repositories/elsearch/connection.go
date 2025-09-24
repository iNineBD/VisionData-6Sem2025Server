package elsearch

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
)

type Config struct {
	Addresses []string
	Username  string
	Password  string

	// Connection settings
	MaxRetries    int
	RetryBackoff  time.Duration
	Timeout       time.Duration
	EnableLogging bool

	// TLS settings
	InsecureSkipVerify bool

	IndexName string
}

type Client struct {
	ES     *elasticsearch.Client
	config *Config
}

// NewClient creates a new Elasticsearch client with the provided configuration
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	// Load from environment variables if not provided in config
	if len(cfg.Addresses) == 0 {
		if url := os.Getenv("ELASTICSEARCH_URL"); url != "" {
			cfg.Addresses = []string{url}
		} else {
			cfg.Addresses = []string{"http://elasticsearch:9200"}
		}
	}

	if cfg.Username == "" {
		if username := os.Getenv("ELASTICSEARCH_USERNAME"); username != "" {
			cfg.Username = username
		} else {
			cfg.Username = "elastic"
		}
	}

	if cfg.Password == "" {
		if password := os.Getenv("ELASTICSEARCH_PASSWORD"); password != "" {
			cfg.Password = password
		}
	}

	// Set defaults
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBackoff == 0 {
		cfg.RetryBackoff = 100 * time.Millisecond
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Configure Elasticsearch client for your setup
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,

		RetryOnStatus: []int{502, 503, 504, 429},
		MaxRetries:    cfg.MaxRetries,
		RetryBackoff: func(i int) time.Duration {
			return cfg.RetryBackoff * time.Duration(i)
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: cfg.Timeout,
			TLSClientConfig: &tls.Config{
				// Since you're using security enabled, but likely with self-signed certs in Docker
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			},
		},
		EnableMetrics:     cfg.EnableLogging,
		EnableDebugLogger: cfg.EnableLogging,
	}

	// Create client
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	client := &Client{
		ES:     es,
		config: cfg,
	}

	// Test connection
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}

	return client, nil
}

// Ping tests the connection to Elasticsearch
func (c *Client) Ping() error {
	res, err := c.ES.Ping()
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed with status: %s", res.Status())
	}

	return nil
}

// Info returns cluster information
func (c *Client) Info() (*esapi.Response, error) {
	return c.ES.Info()
}

// Health returns cluster health information
func (c *Client) Health() (*esapi.Response, error) {
	return c.ES.Cluster.Health()
}

// CreateIndex creates an index with optional mapping
func (c *Client) CreateIndex(indexName string, mapping []byte) error {
	res, err := c.ES.Indices.Create(
		indexName,
		c.ES.Indices.Create.WithBody(bytes.NewReader(mapping)),
		c.ES.Indices.Create.WithPretty(),
	)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()
	if res.IsError() {
		return fmt.Errorf("failed to create index %s: %s", indexName, res.String())
	}

	return nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(indexName string) (bool, error) {
	res, err := c.ES.Indices.Exists([]string{indexName})
	if err != nil {
		return false, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()
	return res.StatusCode == 200, nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(indexName string) error {
	res, err := c.ES.Indices.Delete([]string{indexName})
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", indexName, err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()
	if res.IsError() {
		return fmt.Errorf("failed to delete index %s: %s", indexName, res.String())
	}

	return nil
}
