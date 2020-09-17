package msgraph_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func TestAccUserDataSource_byUserPrincipalName(t *testing.T) {
	dsn := "data.azuread_user_msgraph.test"
	id := tf.AccRandTimeInt()
	password := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSource_byUserPrincipalName(id, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsn, "user_principal_name"),
					resource.TestCheckResourceAttrSet(dsn, "account_enabled"),
					resource.TestCheckResourceAttrSet(dsn, "display_name"),
					resource.TestCheckResourceAttrSet(dsn, "mail_nickname"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byUserPrincipalNameNonexistent(t *testing.T) {
	ri := tf.AccRandTimeInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSource_byUserPrincipalNameNonexistent(ri),
				ExpectError: regexp.MustCompile("no user found with user principal name:"),
			},
		},
	})
}

func TestAccUserDataSource_byObjectId(t *testing.T) {
	dsn := "data.azuread_user_msgraph.test"
	id := tf.AccRandTimeInt()
	password := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSource_byObjectId(id, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsn, "user_principal_name"),
					resource.TestCheckResourceAttrSet(dsn, "account_enabled"),
					resource.TestCheckResourceAttrSet(dsn, "display_name"),
					resource.TestCheckResourceAttrSet(dsn, "mail_nickname"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byObjectIdNonexistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSource_byObjectIdNonexistent(),
				ExpectError: regexp.MustCompile("reading User with ID"),
			},
		},
	})
}

func TestAccUserDataSource_byMailNickname(t *testing.T) {
	dsn := "data.azuread_user_msgraph.test"
	id := tf.AccRandTimeInt()
	password := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSource_byMailNickname(id, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsn, "user_principal_name"),
					resource.TestCheckResourceAttrSet(dsn, "account_enabled"),
					resource.TestCheckResourceAttrSet(dsn, "display_name"),
					resource.TestCheckResourceAttrSet(dsn, "mail_nickname"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byMailNicknameNonexistent(t *testing.T) {
	ri := tf.AccRandTimeInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSource_byMailNicknameNonexistent(ri),
				ExpectError: regexp.MustCompile("no user found with mail nickname:"),
			},
		},
	})
}

func testAccUserDataSource_byUserPrincipalName(id int, password string) string {
	return fmt.Sprintf(`
%s

data "azuread_user_msgraph" "test" {
  user_principal_name = azuread_user_msgraph.test.user_principal_name
}
`, testAccUser_basic(id, password))
}

func testAccUserDataSource_byUserPrincipalNameNonexistent(ri int) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

data "azuread_user_msgraph" "test" {
  user_principal_name = "not-a-real-user-%d${data.azuread_domains.tenant_domain.domains.0.domain_name}"
}
`, ri)
}

func testAccUserDataSource_byObjectId(id int, password string) string {
	return fmt.Sprintf(`
%s

data "azuread_user_msgraph" "test" {
  object_id = azuread_user_msgraph.test.object_id
}
`, testAccUser_basic(id, password))
}

func testAccUserDataSource_byObjectIdNonexistent() string {
	return `
data "azuread_user_msgraph" "test" {
  object_id = "00000000-0000-0000-0000-000000000000"
}
`
}

func testAccUserDataSource_byMailNickname(id int, password string) string {
	return fmt.Sprintf(`
%s

data "azuread_user_msgraph" "test" {
  mail_nickname = azuread_user_msgraph.test.mail_nickname
}
`, testAccUser_basic(id, password))
}

func testAccUserDataSource_byMailNicknameNonexistent(ri int) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

data "azuread_user_msgraph" "test" {
  mail_nickname = "not-a-real-user-%d${data.azuread_domains.tenant_domain.domains.0.domain_name}"
}
`, ri)
}
