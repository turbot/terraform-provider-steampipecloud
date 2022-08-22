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
func TestAccUserWorkspaceSnapshot_Basic(t *testing.T) {
	resourceName := "steampipecloud_workspace_snapshot.snapshot_1"
	workspaceHandle := "workspace" + randomString(3)
	visibility := "workspace"
	updatedVisibility := "anyone_with_link"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceSnapshotConfig(workspaceHandle, visibility),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceSnapshotExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "visibility", visibility),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "data"},
			},
			{
				Config: testAccUserWorkspaceSnapshotUpdateConfig(workspaceHandle, updatedVisibility),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceSnapshotExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "visibility", updatedVisibility),
				),
			},
		},
	})
}

func testAccUserWorkspaceSnapshotConfig(workspaceHandle, visibility string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_snapshot" "snapshot_1" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		data             = jsonencode({
			"action": "execution_complete",
			"dashboard_node": {
				"dashboard": "aws_tags.benchmark.limit",
				"description": "The number of tags on each resource should be monitored to avoid hitting the limit unexpectedly.",
				"name": "aws_tags.benchmark.limit",
				"panel_type": "benchmark",
				"session_id": "0xc001078e40",
				"summary": {
					"status": {
						"alarm": 0,
						"error": 0,
						"info": 0,
						"ok": 40,
						"skip": 0
					}
				},
				"tags": {
					"category": "Tagging",
					"plugin": "aws",
					"service": "AWS",
					"type": "Benchmark"
				},
				"title": "Limit"
			},
			"end_time": "2022-08-11T18:13:45+05:30",
			"inputs": {},
			"layout": {
				"children": [
					{
						"name": "aws_tags.control.accessanalyzer_analyzer_tag_limit",
						"panel_type": "control"
					}
				],
				"name": "aws_tags.benchmark.limit",
				"panel_type": "benchmark"
			},
			"panels": {
				"aws_tags.control.accessanalyzer_analyzer_tag_limit": {
					"data": {
						"columns": [
							{
								"data_type": "TEXT",
								"name": "reason"
							},
							{
								"data_type": "TEXT",
								"name": "resource"
							},
							{
								"data_type": "TEXT",
								"name": "status"
							}
						],
						"rows": []
					},
					"description": "Check if the number of tags on Access Analyzer analyzers do not exceed the limit.",
					"name": "aws_tags.control.accessanalyzer_analyzer_tag_limit",
					"panel_type": "control",
					"properties": {},
					"status": "complete",
					"summary": {
						"alarm": 0,
						"error": 0,
						"info": 0,
						"ok": 0,
						"skip": 0
					},
					"title": "Access Analyzer analyzers should not exceed tag limit"
				}
			},
			"schema_version": "20220614",
			"search_path": [
				"public",
				"aws",
				"steampipecloud",
				"internal"
			],
			"start_time": "2022-08-11T18:13:45+05:30",
			"variables": {
				"aws_tags.var.mandatory_tags": "['Environment','Owner']",
				"aws_tags.var.prohibited_tags": "['Password','Key']",
				"aws_tags.var.tag_limit": "45"
			}
		})
		tags             = jsonencode({
			name: "snapshot_1",
			foo: "bar"
		})
		visibility       = "%s"
	}`, workspaceHandle, visibility)
}

func testAccUserWorkspaceSnapshotUpdateConfig(workspaceHandle, visibility string) string {
	return fmt.Sprintf(`
	provider "steampipecloud" {}

	resource "steampipecloud_workspace" "test_workspace" {
		handle = "%s"
	}
	
	resource "steampipecloud_workspace_snapshot" "snapshot_1" {
		workspace_handle = steampipecloud_workspace.test_workspace.handle
		data             = jsonencode({
			"action": "execution_complete",
			"dashboard_node": {
				"dashboard": "aws_tags.benchmark.limit",
				"description": "The number of tags on each resource should be monitored to avoid hitting the limit unexpectedly.",
				"name": "aws_tags.benchmark.limit",
				"panel_type": "benchmark",
				"session_id": "0xc001078e40",
				"summary": {
					"status": {
						"alarm": 0,
						"error": 0,
						"info": 0,
						"ok": 40,
						"skip": 0
					}
				},
				"tags": {
					"category": "Tagging",
					"plugin": "aws",
					"service": "AWS",
					"type": "Benchmark"
				},
				"title": "Limit"
			},
			"end_time": "2022-08-11T18:13:45+05:30",
			"inputs": {},
			"layout": {
				"children": [
					{
						"name": "aws_tags.control.accessanalyzer_analyzer_tag_limit",
						"panel_type": "control"
					}
				],
				"name": "aws_tags.benchmark.limit",
				"panel_type": "benchmark"
			},
			"panels": {
				"aws_tags.control.accessanalyzer_analyzer_tag_limit": {
					"data": {
						"columns": [
							{
								"data_type": "TEXT",
								"name": "reason"
							},
							{
								"data_type": "TEXT",
								"name": "resource"
							},
							{
								"data_type": "TEXT",
								"name": "status"
							}
						],
						"rows": []
					},
					"description": "Check if the number of tags on Access Analyzer analyzers do not exceed the limit.",
					"name": "aws_tags.control.accessanalyzer_analyzer_tag_limit",
					"panel_type": "control",
					"properties": {},
					"status": "complete",
					"summary": {
						"alarm": 0,
						"error": 0,
						"info": 0,
						"ok": 0,
						"skip": 0
					},
					"title": "Access Analyzer analyzers should not exceed tag limit"
				}
			},
			"schema_version": "20220614",
			"search_path": [
				"public",
				"aws",
				"steampipecloud",
				"internal"
			],
			"start_time": "2022-08-11T18:13:45+05:30",
			"variables": {
				"aws_tags.var.mandatory_tags": "['Environment','Owner']",
				"aws_tags.var.prohibited_tags": "['Password','Key']",
				"aws_tags.var.tag_limit": "45"
			}
		})
		tags             = jsonencode({
			name: "snapshot_1",
			foo: "bar"
		})
		visibility       = "%s"
	}`, workspaceHandle, visibility)
}

func testAccCheckWorkspaceSnapshotExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*SteampipeClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "steampipecloud_workspace_snapshot" {
				continue
			}

			snapshotId := rs.Primary.Attributes["workspace_snapshot_id"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var actorHandle string
				actorHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspaceSnapshots.Get(ctx, actorHandle, workspaceHandle, snapshotId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching snapshot %s in user workspace with handle %s. %s", snapshotId, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceSnapshots.Get(ctx, org, workspaceHandle, snapshotId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching snapshot %s in org workspace with handle %s. %s", snapshotId, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceSnapshotDestroy verifies the snapshot has been deleted from the workspace
func testAccCheckWorkspaceSnapshotDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SteampipeClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "steampipecloud_workspace_snapshot" {
			continue
		}

		// Retrieve workspace handle and snapshot id by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		snapshotId := rs.Primary.Attributes["workspace_snapshot_id"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceSnapshots.Get(ctx, actorHandle, workspaceHandle, snapshotId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceSnapshots.Get(ctx, org, workspaceHandle, snapshotId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Snapshot %s/%s still exists", workspaceHandle, snapshotId)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspaceSnapshotDestroy testAccCheckUserWorkspaceSnapshotDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspaceSnapshotDestroy testAccCheckUserWorkspaceSnapshotDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
