package steampipecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/turbot/go-kit/helpers"
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

// apply standard formatting to a json string by unmarshalling into a map then marshalling back to JSON
func formatJson(body string) (string, map[string]interface{}) {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, []byte(body))
	if err != nil {
		return body, nil
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(buffer.Bytes(), &data); err != nil {
		// ignore error and just return original body
		return body, nil
	}
	for k, v := range data {
		if helpers.StringSliceContains([]string{"credentials"}, k) {
			bufferCredentials := new(bytes.Buffer)
			if compactErr := json.Compact(bufferCredentials, []byte(v.(string))); compactErr != nil {
				continue
			}
			data[k] = bufferCredentials.String()
		} else if k == "private_key" {
			privateKey := v.(string)
			data[k] = strings.ReplaceAll(privateKey, "\r\n", "\\n")
		}
	}

	body, err = mapToJsonString(data)
	if err != nil {
		// ignore error and just return original body
		return body, data
	}
	return body, data
}

func mapToJsonString(data map[string]interface{}) (string, error) {
	dataBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return "", err
	}
	jsonData := string(dataBytes)
	return jsonData, nil
}

func JsonStringToMap(dataString string) (map[string]interface{}, error) {
	var data = make(map[string]interface{})
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		return nil, err
	}
	return data, nil
}
