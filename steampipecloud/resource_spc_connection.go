package steampipe

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	openapiclient "github.com/turbot/steampipe-cloud-sdk-go"
)

func resourceSteampipeCloudConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSteampipeCloudConnectionCreate,
		Read:   resourceSteampipeCloudConnectionRead,
		Delete: resourceSteampipeCloudConnectionDelete,
		Exists: resourceSteampipeCloudConnectionExists,
		Importer: &schema.ResourceImporter{
			State: resourceSteampipeCloudConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			// aka of the parent resource
			"plugin": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSteampipeCloudConnectionExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	return false, nil
}

func resourceSteampipeCloudConnectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSteampipeCloudConnectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceSteampipeCloudConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	req := openapiclient.TypesCreateConnectionRequest{
		Handle: "subhajitintg2021",
		Plugin: "rss",
	}
	resp, r, err := client.OrgsConnectionsApi.OrgOrgHandleConnPost(context.Background(), "netaji").Request(req).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspacePost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// data := *resp.Items
	d.Set("plugin", resp.Plugin)

	return nil
}

func resourceSteampipeCloudConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	resp, r, err := client.OrgsConnectionsApi.OrgOrgHandleConnGet(context.Background(), "netaji").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleGet`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	data := *resp.Items

	for _, i := range data {
		if *i.Plugin == "rss" {
			d.Set("plugin", i.Plugin)
		}
	}

	return nil
}

func resourceSteampipeCloudConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*openapiclient.APIClient)

	_, r, err := client.OrgsConnectionsApi.OrgOrgHandleConnConnHandleDelete(context.Background(), "netaji", "rss").Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UserUserHandleWorkspaceWorkspaceHandleDelete`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return nil
}
