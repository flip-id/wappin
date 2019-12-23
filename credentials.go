package wappin

import (
	"encoding/base64"
	"os"
	"strings"
)

var baseUrl = os.Getenv("WAPPIN_BASE_URL")

type Credentials struct {
	clientKey, clientSecret string
}

var credentials Credentials

func init() {
	setCredentials()
}

func setCredentials() {
	envPrefix := "WAPPIN_"

	credentials = Credentials{
		os.Getenv(envPrefix + "CLIENT_KEY"),
		os.Getenv(envPrefix + "CLIENT_SECRET"),
	}
}

// Get basic auth
func (c Credentials) getBasicAuth(clientId string) string {
	basicAuth := strings.Join([]string{clientId, c.clientSecret}, ":")

	return base64.StdEncoding.EncodeToString([]byte(basicAuth))
}
