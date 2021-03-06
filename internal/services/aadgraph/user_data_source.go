package aadgraph

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph/graph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func userData() *schema.Resource {
	return &schema.Resource{
		Read: userDataRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.UUID,
				ConflictsWith: []string{"user_principal_name"},
			},

			"user_principal_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.NoEmptyStrings,
				ConflictsWith: []string{"object_id"},
			},

			"mail_nickname": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.NoEmptyStrings,
				ConflictsWith: []string{"object_id", "user_principal_name"},
			},

			"account_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"immutable_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mail": {
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
		},
	}
}

func userDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).AadGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	var user graphrbac.User

	if upn, ok := d.Get("user_principal_name").(string); ok && upn != "" {
		resp, err := client.Get(ctx, upn)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("User not found with UPN: %q", upn)
			}
			return fmt.Errorf("retrieving User with ID %q: %+v", upn, err)
		}
		user = resp
	} else if oId, ok := d.Get("object_id").(string); ok && oId != "" {
		u, err := graph.UserGetByObjectId(ctx, client, oId)
		if err != nil {
			return fmt.Errorf("finding User with object ID %q: %+v", oId, err)
		}
		if u == nil {
			return fmt.Errorf("User not found with object ID: %q", oId)
		}
		user = *u
	} else if mailNickname, ok := d.Get("mail_nickname").(string); ok && mailNickname != "" {
		u, err := graph.UserGetByMailNickname(ctx, client, mailNickname)
		if err != nil {
			return fmt.Errorf("finding User with email alias %q: %+v", mailNickname, err)
		}
		if u == nil {
			return fmt.Errorf("User not found with email alias: %q", mailNickname)
		}
		user = *u
	} else {
		return fmt.Errorf("one of `object_id`, `user_principal_name` and `mail_nickname` must be supplied")
	}

	if user.ObjectID == nil {
		return fmt.Errorf("User objectId is nil")
	}
	d.SetId(*user.ObjectID)

	d.Set("object_id", user.ObjectID)
	d.Set("user_principal_name", user.UserPrincipalName)
	d.Set("account_enabled", user.AccountEnabled)
	d.Set("display_name", user.DisplayName)
	d.Set("immutable_id", user.ImmutableID)
	d.Set("mail", user.Mail)
	d.Set("mail_nickname", user.MailNickname)
	d.Set("usage_location", user.UsageLocation)

	d.Set("onpremises_sam_account_name", user.AdditionalProperties["onPremisesSamAccountName"])
	d.Set("onpremises_user_principal_name", user.AdditionalProperties["onPremisesUserPrincipalName"])

	return nil
}
