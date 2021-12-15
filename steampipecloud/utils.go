package steampipecloud

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
)

// isUserConnection:: Check if the connection is scoped on an user or a specific organization
func isUserConnection(client *SteampipeClient) (ok bool, orgHandle string) {
	ok = true
	if client.Config != nil {
		if client.Config.Organization != "" {
			orgHandle = client.Config.Organization
			ok = false
		}
	}
	return
}

// helper functions
func getUserHandler(ctx context.Context, client *SteampipeClient) (string, *http.Response, error) {
	resp, r, err := client.APIClient.Actors.Get(ctx).Execute()
	if err != nil {
		return "", r, err
	}
	return resp.Handle, r, nil
}

// Decode response body
func decodeResponse(r *http.Response) string {
	var errBody interface{}
	_ = json.NewDecoder(r.Body).Decode(&errBody)
	defer r.Body.Close()

	resp, _ := json.Marshal(errBody)
	return string(resp)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

// randomString:: To generate random names for handle for testing
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
