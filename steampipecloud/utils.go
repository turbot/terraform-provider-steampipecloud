package steampipecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	steampipe "github.com/turbot/steampipe-cloud-sdk-go"
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

func getWorkspaceDetails(ctx context.Context, client *SteampipeClient, d *schema.ResourceData) (*steampipe.Workspace, *http.Response, error) {
	var resp steampipe.Workspace
	var r *http.Response
	var err error
	// Get the workspace handle information
	workspaceHandle := d.Get("workspace_handle").(string)
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return nil, r, err
		}
		resp, r, err = client.APIClient.UserWorkspaces.Get(ctx, actorHandle, workspaceHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
	}

	if err != nil {
		return nil, r, err
	}
	return &resp, r, nil
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

func TestJSONFieldEqual(t *testing.T, resourceName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fieldValue := s.RootModule().Resources[resourceName].Primary.Attributes[key]
		require.JSONEq(t, fieldValue, value)
		return nil
	}
}

func TestArrayEqual(t *testing.T, resourceName, key, values string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Unmarshal the values passed as a string into a string array
		var matcherItems []string
		err := json.Unmarshal([]byte(values), &matcherItems)
		if err != nil {
			return err
		}

		// Get the number of expected items in the primary attribute
		noOfExpectedItems, err := strconv.Atoi(s.RootModule().Resources[resourceName].Primary.Attributes[fmt.Sprintf("%s.#", key)])
		if err != nil {
			return err
		}

		// If the number of items expected and the no of items present do not match throw an error
		if len(matcherItems) != noOfExpectedItems {
			return fmt.Errorf("number of expected items do not match for attribute %s", key)
		}

		for i := 0; i < noOfExpectedItems; i++ {
			if matcherItems[i] != s.RootModule().Resources[resourceName].Primary.Attributes[fmt.Sprintf("%s.%d", key, i)] {
				return fmt.Errorf("mismatch while matching item no %d for attribute %s", i, key)
			}
		}
		return nil
	}
}

func convertToStringArray(data []interface{}) ([]string, error) {
	strArray := make([]string, len(data))
	for index, value := range data {
		if s, ok := value.(string); ok {
			strArray[index] = s
		} else {
			return nil, fmt.Errorf("invalid value entered in string array")
		}
	}
	return strArray, nil
}
