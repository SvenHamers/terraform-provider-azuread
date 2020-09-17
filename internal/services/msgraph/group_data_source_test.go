package msgraph_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
)

func TestAccGroupDataSource_byName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_group_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSource_name(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
				),
			},
		},
	})
}

func TestAccGroupDataSource_byCaseInsensitiveName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_group_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSource_caseInsensitiveName(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
				),
			},
		},
	})
}

func TestAccGroupDataSource_byObjectId(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_group_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSource_objectId(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
				),
			},
		},
	})
}

func TestAccGroupDataSource_members(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSource_members(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "members.#", "3"),
				),
			},
		},
	})
}

func TestAccGroupDataSource_owners(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSource_owners(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "owners.#", "3"),
				),
			},
		},
	})
}

func testAccGroupDataSource_name(id int) string {
	return fmt.Sprintf(`
%s

data "azuread_group_msgraph" "test" {
  display_name = azuread_group_msgraph.test.display_name
}
`, testAccGroup_basic(id))
}

func testAccGroupDataSource_caseInsensitiveName(id int) string {
	return fmt.Sprintf(`
%s
data "azuread_group_msgraph" "test" {
  display_name = upper(azuread_group_msgraph.test.display_name)
}
`, testAccGroup_basic(id))
}

func testAccGroupDataSource_objectId(id int) string {
	return fmt.Sprintf(`
%s

data "azuread_group_msgraph" "test" {
  object_id = azuread_group_msgraph.test.object_id
}
`, testAccGroup_basic(id))
}

func testAccGroupDataSource_members(id int, password string) string {
	return fmt.Sprintf(`
%s

data "azuread_group_msgraph" "test" {
  object_id = azuread_group_msgraph.test.object_id
}
`, testAccGroupWithThreeMembers(id, password))
}

func testAccGroupDataSource_owners(id int, password string) string {
	return fmt.Sprintf(`
%s

data "azuread_group_msgraph" "test" {
  object_id = azuread_group_msgraph.test.object_id
}
`, testAccGroupWithThreeOwners(id, password))
}
