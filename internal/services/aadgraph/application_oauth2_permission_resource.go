package aadgraph

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph/graph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func applicationOAuth2PermissionResource() *schema.Resource {
	return &schema.Resource{
		Create: applicationOAuth2PermissionResourceCreateUpdate,
		Update: applicationOAuth2PermissionResourceCreateUpdate,
		Read:   applicationOAuth2PermissionResourceRead,
		Delete: applicationOAuth2PermissionResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := graph.ParseOAuth2PermissionId(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"application_object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},

			"admin_consent_description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"admin_consent_display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"permission_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"Admin", "User"},
					false,
				),
			},

			"user_consent_description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"user_consent_display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},
		},
	}
}

func applicationOAuth2PermissionResourceCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).AadGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	objectId := d.Get("application_object_id").(string)

	// errors should be handled by the validation
	var permissionId string

	if v, ok := d.GetOk("permission_id"); ok {
		permissionId = v.(string)
	} else {
		pid, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("generating OAuth2 Permission for Object ID %q: %+v", objectId, err)
		}
		permissionId = pid
	}

	permission := graphrbac.OAuth2Permission{
		AdminConsentDescription: utils.String(d.Get("admin_consent_description").(string)),
		AdminConsentDisplayName: utils.String(d.Get("admin_consent_display_name").(string)),
		ID:                      utils.String(permissionId),
		IsEnabled:               utils.Bool(d.Get("is_enabled").(bool)),
		Type:                    utils.String(d.Get("type").(string)),
		UserConsentDescription:  utils.String(d.Get("user_consent_description").(string)),
		UserConsentDisplayName:  utils.String(d.Get("user_consent_display_name").(string)),
		Value:                   utils.String(d.Get("value").(string)),
	}

	id := graph.OAuth2PermissionIdFrom(objectId, *permission.ID)

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	// ensure the Application Object exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		if utils.ResponseWasNotFound(app.Response) {
			return fmt.Errorf("Application with ID %q was not found", id.ObjectId)
		}
		return fmt.Errorf("retrieving Application ID %q: %+v", id.ObjectId, err)
	}

	var newPermissions *[]graphrbac.OAuth2Permission

	if d.IsNewResource() {
		newPermissions, err = graph.OAuth2PermissionAdd(app.Oauth2Permissions, &permission)
		if err != nil {
			if _, ok := err.(*graph.AlreadyExistsError); ok {
				return tf.ImportAsExistsError("azuread_application_oauth2_permission", id.String())
			}
			return fmt.Errorf("adding OAuth2 Permission: %+v", err)
		}
	} else {
		if existing, _ := graph.OAuth2PermissionFindById(app, id.PermissionId); existing == nil {
			return fmt.Errorf("OAuth2 Permission with ID %q was not found for Application %q", id.PermissionId, id.ObjectId)
		}

		newPermissions, err = graph.OAuth2PermissionUpdate(app.Oauth2Permissions, &permission)
		if err != nil {
			return fmt.Errorf("updating OAuth2 Permission: %s", err)
		}
	}

	properties := graphrbac.ApplicationUpdateParameters{
		Oauth2Permissions: newPermissions,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return fmt.Errorf("patching Application with ID %q: %+v", id.ObjectId, err)
	}

	d.SetId(id.String())

	return applicationOAuth2PermissionResourceRead(d, meta)
}

func applicationOAuth2PermissionResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).AadGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	id, err := graph.ParseOAuth2PermissionId(d.Id())
	if err != nil {
		return fmt.Errorf("parsing OAuth2 Permission ID: %v", err)
	}

	// ensure the Application Object exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Application has been removed - skip it
		if utils.ResponseWasNotFound(app.Response) {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Application ID %q: %+v", id.ObjectId, err)
	}

	permission, err := graph.OAuth2PermissionFindById(app, id.PermissionId)
	if err != nil {
		return fmt.Errorf("identifying OAuth2 Permission: %s", err)
	}

	if permission == nil {
		log.Printf("[DEBUG] OAuth2 Permission %q (ID %q) was not found - removing from state!", id.PermissionId, id.ObjectId)
		d.SetId("")
		return nil
	}

	d.Set("application_object_id", id.ObjectId)
	d.Set("permission_id", id.PermissionId)
	d.Set("admin_consent_description", permission.AdminConsentDescription)
	d.Set("admin_consent_display_name", permission.AdminConsentDisplayName)
	d.Set("is_enabled", permission.IsEnabled)
	d.Set("type", permission.Type)
	d.Set("user_consent_description", permission.UserConsentDescription)
	d.Set("user_consent_display_name", permission.UserConsentDisplayName)
	d.Set("value", permission.Value)

	return nil
}

func applicationOAuth2PermissionResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).AadGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	id, err := graph.ParseOAuth2PermissionId(d.Id())
	if err != nil {
		return fmt.Errorf("parsing OAuth2 Permission ID: %v", err)
	}

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	// ensure the parent Application exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Application has been removed - skip it
		if utils.ResponseWasNotFound(app.Response) {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			return nil
		}
		return fmt.Errorf("retrieving Application ID %q: %+v", id.ObjectId, err)
	}

	var newPermissions *[]graphrbac.OAuth2Permission

	log.Printf("[DEBUG] Disabling OAuth2 Permission %q for Application %q prior to removal", id.PermissionId, id.ObjectId)
	newPermissions, err = graph.OAuth2PermissionResultDisableById(app.Oauth2Permissions, id.PermissionId)
	if err != nil {
		return fmt.Errorf("could not disable OAuth2 Permission prior to removal: %s", err)
	}

	properties := graphrbac.ApplicationUpdateParameters{
		Oauth2Permissions: newPermissions,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return fmt.Errorf("patching Application with ID %q: %+v", id.ObjectId, err)
	}

	log.Printf("[DEBUG] Removing OAuth2 Permission %q for Application %q", id.PermissionId, id.ObjectId)
	newPermissions, err = graph.OAuth2PermissionResultRemoveById(app.Oauth2Permissions, id.PermissionId)
	if err != nil {
		return fmt.Errorf("could not remove OAuth2 Permission: %s", err)
	}

	properties = graphrbac.ApplicationUpdateParameters{
		Oauth2Permissions: newPermissions,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return fmt.Errorf("patching Application with ID %q: %+v", id.ObjectId, err)
	}

	return nil
}
