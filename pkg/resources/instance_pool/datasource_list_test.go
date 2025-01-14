package instance_pool_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/exoscale/terraform-provider-exoscale/pkg/testutils"
)

var (
	dsListDiskSize     = "10"
	dsListInstanceType = "standard.tiny"
	dsListKeyPair      = acctest.RandomWithPrefix(testutils.Prefix)
	dsListName         = acctest.RandomWithPrefix(testutils.Prefix)
)

var dsListConfig = fmt.Sprintf(`
locals {
  zone = "%s"
	instance_type = "%s"
	disk_size = "%s"
}
resource "exoscale_ssh_keypair" "test" {
  name = "%s"
}
data "exoscale_compute_template" "ubuntu" {
  zone = local.zone
  name = "Linux Ubuntu 20.04 LTS 64-bit"
}
resource "exoscale_instance_pool" "test1" {
  zone = local.zone
  name = "%s"
  template_id = data.exoscale_compute_template.ubuntu.id
  instance_type = local.instance_type
  size = 1
  disk_size = local.disk_size
  key_pair = exoscale_ssh_keypair.test.name
}
resource "exoscale_instance_pool" "test2" {
  zone = local.zone
  name = "%s"
  template_id = data.exoscale_compute_template.ubuntu.id
  instance_type = local.instance_type
  size = 1
  disk_size = local.disk_size
  key_pair = exoscale_ssh_keypair.test.name
  labels = { test="test"}
}`,
	testutils.TestZoneName,
	dsListInstanceType,
	dsListDiskSize,
	dsListKeyPair,
	dsListName+"_1",
	dsListName+"_2",
)

func testListDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testutils.AccPreCheck(t) },
		ProviderFactories: testutils.Providers(),
		Steps: []resource.TestStep{
			{
				Config: dsListConfig,
			},
			{
				Config: fmt.Sprintf(`
%s
data "exoscale_instance_pool_list" "test" {
  # we omit the zone to trigger an error as the zone attribute must be mandatory.
}
`,
					dsListConfig,
				),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config: fmt.Sprintf(`
			%s
			data "exoscale_instance_pool_list" "test" {
			  zone = local.zone
			}
			`,
					dsListConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					dsCheckListAttrs("data.exoscale_instance_pool_list.test", testutils.TestAttrs{
						"pools.#":             testutils.ValidateString("2"),
						"pools.0.id":          validation.ToDiagFunc(validation.NoZeroValues),
						"pools.0.instances.#": testutils.ValidateString("1"),
						"pools.1.id":          validation.ToDiagFunc(validation.NoZeroValues),
						"pools.1.instances.#": testutils.ValidateString("1"),
					}),
				),
			},
		},
	})
}

func dsCheckListAttrs(ds string, expected testutils.TestAttrs) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, res := range s.RootModule().Resources {
			if name == ds {
				return testutils.CheckResourceAttributes(expected, res.Primary.Attributes)
			}
		}

		return errors.New("exoscale_instance_pool_list data source not found in the state")
	}
}
