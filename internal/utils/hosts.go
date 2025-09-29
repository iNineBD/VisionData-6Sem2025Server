package utils

import (
	"os"

	"github.com/gin-gonic/gin"
)

// GetCurrentProtocolAndHost returns the current protocol and host
func GetCurrentProtocolAndHost(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	host := c.Request.Host

	return protocol + "://" + host
}

// GetPort returns the port from the environment variable PORT
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// GetCertFiles returns the certificate files from the environment variables
func GetCertFiles() (string, string) {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	return certFile, keyFile
}
