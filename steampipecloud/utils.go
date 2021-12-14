package steampipecloud

import (
	"context"
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
