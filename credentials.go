package wappin

import (
	"encoding/base64"
	"os"
	"strings"
)

var baseUrl = os.Getenv("WAPPIN_BASE_URL")
var clientId = os.Getenv("WAPPIN_CLIENT_ID")

// Get basic auth
func getBasicAuth(clientSecret string) string {
	basicAuth := strings.Join([]string{clientId, clientSecret}, ":")

	return base64.StdEncoding.EncodeToString([]byte(basicAuth))
}
