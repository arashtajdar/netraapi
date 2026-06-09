package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

func getCDNSecret() string {
	secret := os.Getenv("CDN_SECRET_KEY")
	if secret == "" {
		return "development_secret_do_not_use_in_prod"
	}
	return secret
}

// SignURL takes a raw media URL, appends an expiration timestamp and a cryptographic HMAC token.
func SignURL(rawURL string) string {
	if rawURL == "" || !strings.HasPrefix(rawURL, "http") {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Expiration time: current time + 2 hours
	expires := time.Now().Add(2 * time.Hour).Unix()
	
	// The token is an HMAC of the URL path and the expiration time
	message := fmt.Sprintf("%s-%d", u.Path, expires)
	
	mac := hmac.New(sha256.New, []byte(getCDNSecret()))
	mac.Write([]byte(message))
	token := hex.EncodeToString(mac.Sum(nil))

	q := u.Query()
	q.Set("expires", fmt.Sprintf("%d", expires))
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}

// SignVideoSources takes a JSON array of video sources, signs their URLs, and returns the modified JSON.
func SignVideoSources(rawJSON []byte) []byte {
	if rawJSON == nil || len(rawJSON) == 0 || string(rawJSON) == "[]" {
		return rawJSON
	}

	var sources []map[string]interface{}
	if err := json.Unmarshal(rawJSON, &sources); err != nil {
		return rawJSON
	}

	for i, source := range sources {
		if urlVal, ok := source["url"].(string); ok {
			sources[i]["url"] = SignURL(urlVal)
		}
	}

	signedJSON, err := json.Marshal(sources)
	if err != nil {
		return rawJSON
	}

	return signedJSON
}
