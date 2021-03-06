package steampipecloud

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// isUserConnection:: Check if the connection is scoped on an user or a specific organization
func isUserConnection(d *schema.ResourceData) (isUser bool, orgHandle string) {
	isUser = true

	if val, ok := d.GetOk("organization"); ok {
		orgHandle = val.(string)
		isUser = false
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

func mapToJSONString(data map[string]interface{}) (string, error) {
	dataBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return "", err
	}
	jsonData := string(dataBytes)
	return jsonData, nil
}

func JSONStringToInterface(dataString string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// apply standard formatting to a json string by unmarshalling into a map then marshalling back to JSON
func FormatJson(body interface{}) string {
	var raw []byte
	var err error
	if raw, err = marshallInterfaceToByteArray(body); err != nil {
		// ignore error and just return original body
		return ""
	}

	return string(raw)
}

func marshallInterfaceToByteArray(data interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return dataBytes, nil
}
