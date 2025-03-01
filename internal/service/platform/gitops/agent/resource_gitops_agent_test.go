package agent_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/antihax/optional"
	"github.com/harness/harness-go-sdk/harness/nextgen"
	"github.com/harness/harness-go-sdk/harness/utils"
	"github.com/harness/terraform-provider-harness/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceGitopsAgent(t *testing.T) {
	id := fmt.Sprintf("%s_%s", t.Name(), utils.RandStringBytes(5))
	accountId := os.Getenv("HARNESS_ACCOUNT_ID")
	resourceName := "harness_platform_gitops_agent.test"
	agentName := id
	namespace := "terraform-test"
	updatedNamespace := namespace + "-updated"
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccResourceGitopsAgentDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGitopsAgent(id, accountId, agentName, namespace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", agentName),
				),
			},
			{
				Config: testAccResourceGitopsAgent(id, accountId, agentName, updatedNamespace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "metadata.0.namespace", updatedNamespace),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"account_id", "type"},
				ImportStateIdFunc:       acctest.ProjectResourceImportStateIdFunc(resourceName),
			},
		},
	})

}

func testAccGetAgent(resourceName string, state *terraform.State) (*nextgen.V1Agent, error) {
	r := acctest.TestAccGetResource(resourceName, state)
	c, ctx := acctest.TestAccGetPlatformClientWithContext()
	// ctx = context.WithValue(ctx, nextgen.ContextAccessToken, hh.EnvVars.BearerToken.Get())

	agentIdentifier := r.Primary.Attributes["identifier"]

	resp, _, err := c.AgentApi.AgentServiceForServerGet(ctx, agentIdentifier, c.AccountId, &nextgen.AgentsApiAgentServiceForServerGetOpts{
		OrgIdentifier:     optional.NewString(r.Primary.Attributes["org_identifier"]),
		ProjectIdentifier: optional.NewString(r.Primary.Attributes["project_identifier"]),
	})

	if err != nil {
		return nil, err
	}

	if resp.Type_ == nil {
		return nil, nil
	}

	return &resp, nil
}

func testAccResourceGitopsAgentDestroy(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		agent, _ := testAccGetAgent(resourceName, state)
		if agent != nil {
			return fmt.Errorf("Found Agent: %s", agent.Identifier)
		}
		return nil
	}

}

func testAccResourceGitopsAgent(agentId string, accountId string, agentName string, namespace string) string {
	return fmt.Sprintf(`
		resource "harness_platform_organization" "test" {
			identifier = "%[1]s"
			name = "%[2]s"
		}

		resource "harness_platform_project" "test" {
			identifier = "%[1]s"
			name = "%[2]s"
			org_id = harness_platform_organization.test.id
		}
		resource "harness_platform_gitops_agent" "test" {
			identifier = "%[1]s"
			account_id = "%[2]s"
			project_id = harness_platform_project.test.id
			org_id = harness_platform_organization.test.id
			name = "%[3]s"
			type = "CONNECTED_ARGO_PROVIDER"
			metadata {
        namespace = "%[4]s"
        high_availability = true
    	}
		}
		`, agentId, accountId, agentName, namespace)
}
