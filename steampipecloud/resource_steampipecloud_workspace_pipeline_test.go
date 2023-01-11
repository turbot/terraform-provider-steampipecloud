package steampipecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccUserWorkspacePipeline_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_pipeline.pipeline_1"
	workspaceHandle := "workspace" + randomString(3)
	title := "Daily CIS Job"
	pipeline := "pipeline.save_snapshot"
	frequency := `
		{
			"type": "interval",
			"schedule": "daily"
		}
	`
	args := `
		{
			"resource": "aws_compliance.benchmark.cis_v140",
			"identity_type": "user",
			"identity_handle": "testuser",
			"workspace_handle": "dev",
			"inputs": {},
			"tags": {
				"series": "daily_cis"
			}
		}
	`
	tags := `
		{
			"name": "pipeline_1",
			"foo": "bar"
		}
	`
	updatedFrequency := `
		{
			"type": "interval",
			"schedule": "hourly"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspacePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspacePipelineConfig(workspaceHandle, title, pipeline, frequency, args, tags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspacePipelineExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "pipeline", pipeline),
					TestJSONFieldEqual(t, resourceName, "frequency", frequency),
					TestJSONFieldEqual(t, resourceName, "args", args),
					TestJSONFieldEqual(t, resourceName, "tags", tags),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "args", "frequency", "tags"},
			},
			{
				Config: testAccUserWorkspacePipelineUpdateConfig(workspaceHandle, title, pipeline, updatedFrequency, args, tags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspacePipelineExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "pipeline", pipeline),
					TestJSONFieldEqual(t, resourceName, "frequency", updatedFrequency),
					TestJSONFieldEqual(t, resourceName, "args", args),
					TestJSONFieldEqual(t, resourceName, "tags", tags),
				),
			},
		},
	})
}

func testAccUserWorkspacePipelineConfig(workspaceHandle, title, pipeline, frequency, args, tags string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_pipeline" "pipeline_1" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		title            = "%s"
		pipeline         = "%s"
		frequency        = jsonencode(%s)
		args             = jsonencode(%s)
		tags             = jsonencode(%s)
	}`, workspaceHandle, title, pipeline, frequency, args, tags)
}

func testAccUserWorkspacePipelineUpdateConfig(workspaceHandle, title, pipeline, frequency, args, tags string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_pipeline" "pipeline_1" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		title            = "%s"
		pipeline         = "%s"
		frequency        = jsonencode(%s)
		args             = jsonencode(%s)
		tags             = jsonencode(%s)
	}`, workspaceHandle, title, pipeline, frequency, args, tags)
}

func testAccCheckWorkspacePipelineExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "steampipecloud_workspace_pipeline" {
				continue
			}

			pipelineId := rs.Primary.Attributes["workspace_pipeline_id"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspacePipelines.Get(ctx, userHandle, workspaceHandle, pipelineId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in user workspace with handle %s. %s", pipelineId, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspacePipelines.Get(ctx, org, workspaceHandle, pipelineId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in org workspace with handle %s. %s", pipelineId, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspacePipelineDestroy verifies the pipeline has been deleted from the workspace
func testAccCheckWorkspacePipelineDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace_pipeline" {
			continue
		}

		// Retrieve workspace handle and pipeline id by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		pipelineId := rs.Primary.Attributes["workspace_pipeline_id"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspacePipelines.Get(ctx, userHandle, workspaceHandle, pipelineId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspacePipelines.Get(ctx, org, workspaceHandle, pipelineId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Pipeline %s/%s still exists", workspaceHandle, pipelineId)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspacePipelineDestroy testAccCheckUserWorkspacePipelineDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspacePipelineDestroy testAccCheckUserWorkspacePipelineDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
