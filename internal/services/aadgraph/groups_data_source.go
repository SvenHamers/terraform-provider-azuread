package aadgraph

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph/graph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func groupsData() *schema.Resource {
	return &schema.Resource{
		Read: GroupsDataRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_ids": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"names", "object_ids"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"names": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"names", "object_ids"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.NoEmptyStrings,
				},
			},
		},
	}
}

func GroupsDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).AadGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	var groups []graphrbac.ADGroup
	expectedCount := 0

	if names, ok := d.Get("names").([]interface{}); ok && len(names) > 0 {
		expectedCount = len(names)
		for _, v := range names {
			g, err := graph.GroupGetByDisplayName(ctx, client, v.(string))
			if err != nil {
				return fmt.Errorf("finding Group with display name %q: %+v", v.(string), err)
			}
			groups = append(groups, *g)
		}
	} else if oids, ok := d.Get("object_ids").([]interface{}); ok && len(oids) > 0 {
		expectedCount = len(oids)
		for _, v := range oids {
			resp, err := client.Get(ctx, v.(string))
			if err != nil {
				return fmt.Errorf("retrieving Group with ID %q: %+v", v.(string), err)
			}

			groups = append(groups, resp)
		}
	}

	if len(groups) != expectedCount {
		return fmt.Errorf("Unexpected number of groups returned (%d != %d)", len(groups), expectedCount)
	}

	names := make([]string, 0, len(groups))
	oids := make([]string, 0, len(groups))
	for _, u := range groups {
		if u.ObjectID == nil || u.DisplayName == nil {
			return fmt.Errorf("Group with nil ObjectID or DisplayName was returned: %v", u)
		}

		oids = append(oids, *u.ObjectID)
		names = append(names, *u.DisplayName)
	}

	h := sha1.New()
	if _, err := h.Write([]byte(strings.Join(names, "-"))); err != nil {
		return fmt.Errorf("Unable to compute hash for names: %v", err)
	}

	d.SetId("groups#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	d.Set("object_ids", oids)
	d.Set("names", names)
	return nil
}
