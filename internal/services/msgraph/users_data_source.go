package msgraph

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func UsersData() *schema.Resource {
	return &schema.Resource{
		Read: usersDataRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_ids": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"mail_nicknames", "object_ids", "user_principal_names"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"mail_nicknames": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"mail_nicknames", "object_ids", "user_principal_names"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.NoEmptyStrings,
				},
			},

			"user_principal_names": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"mail_nicknames", "object_ids", "user_principal_names"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.NoEmptyStrings,
				},
			},

			"ignore_missing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"account_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"mail": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"mail_nickname": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"onpremises_immutable_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"onpremises_sam_account_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"onpremises_user_principal_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"usage_location": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"user_principal_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func usersDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	var users []models.User
	expectedCount := 0
	ignoreMissing := d.Get("ignore_missing").(bool)

	if objectIds, ok := d.Get("object_ids").([]interface{}); ok && len(objectIds) > 0 {
		expectedCount = len(objectIds)
		for _, v := range objectIds {
			user, err := client.Get(ctx, v.(string))
			if err != nil {
				// TODO: implement ignore_missing
				if ignoreMissing {
					continue
				}
				return fmt.Errorf("finding User with ID %q: %+v", v.(string), err)
			}
			users = append(users, *user)
		}
	} else if mailNicknames, ok := d.Get("mail_nicknames").([]interface{}); ok && len(mailNicknames) > 0 {
		expectedCount = len(mailNicknames)
		for _, v := range mailNicknames {
			filter := fmt.Sprintf("mailNickname eq '%s'", v.(string))
			result, err := client.List(ctx, filter)
			if err != nil {
				// TODO: implement ignore_missing
				if ignoreMissing {
					continue
				}
				return fmt.Errorf("finding User with email alias %q: %+v", v.(string), err)
			}
			if len(*result) == 0 {
				if ignoreMissing {
					continue
				}
				return fmt.Errorf("no user found with mail nickname: %q", v.(string))
			}
			users = append(users, (*result)[0])
		}
	} else if upns, ok := d.Get("user_principal_names").([]interface{}); ok && len(upns) > 0 {
		expectedCount = len(upns)
		for _, v := range upns {
			filter := fmt.Sprintf("userPrincipalName eq '%s'", v.(string))
			result, err := client.List(ctx, filter)
			if err != nil {
				// TODO: implement ignore_missing
				if ignoreMissing {
					continue
				}
				return fmt.Errorf("identifying User with user principal name %q: %+v", v.(string), err)
			}
			if len(*result) == 0 {
				if ignoreMissing {
					continue
				}
				return fmt.Errorf("no user found with user principal name: %q", v.(string))
			}
			users = append(users, (*result)[0])
		}
	}

	if !ignoreMissing && len(users) != expectedCount {
		return fmt.Errorf("unexpected number of users returned (%d != %d)", len(users), expectedCount)
	}

	upns := make([]string, 0, len(users))
	oids := make([]string, 0, len(users))
	mailNicknames := make([]string, 0, len(users))
	userList := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		if u.ID == nil || u.UserPrincipalName == nil {
			return fmt.Errorf("user with null ID or UPN was found: %v", u)
		}

		oids = append(oids, *u.ID)
		upns = append(upns, *u.UserPrincipalName)
		mailNicknames = append(mailNicknames, *u.MailNickname)

		user := make(map[string]interface{})
		user["account_enabled"] = u.AccountEnabled
		user["display_name"] = u.DisplayName
		user["mail"] = u.Mail
		user["mail_nickname"] = u.MailNickname
		user["object_id"] = u.ID
		user["onpremises_immutable_id"] = u.OnPremisesImmutableId
		user["onpremises_sam_account_name"] = u.OnPremisesSamAccountName
		user["onpremises_user_principal_name"] = u.OnPremisesUserPrincipalName
		user["usage_location"] = u.UsageLocation
		user["user_principal_name"] = u.UserPrincipalName
		userList = append(userList, user)
	}

	h := sha1.New()
	if _, err := h.Write([]byte(strings.Join(upns, "-"))); err != nil {
		return fmt.Errorf("unable to compute hash for UPNs: %v", err)
	}

	d.SetId("users#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	d.Set("object_ids", oids)
	d.Set("user_principal_names", upns)
	d.Set("mail_nicknames", mailNicknames)
	d.Set("users", userList)

	return nil
}
