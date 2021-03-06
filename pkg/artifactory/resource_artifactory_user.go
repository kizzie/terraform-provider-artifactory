package artifactory

import (
	"math/rand"

	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

const randomPasswordLength = 16

func resourceArtifactoryUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"profile_updatable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"disable_ui_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"internal_password_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
		},
	}
}

func unmarshalUser(s *schema.ResourceData) *artifactory.User {
	d := &ResourceData{s}
	user := new(artifactory.User)

	user.Name = d.GetStringRef("name")
	user.Email = d.GetStringRef("email")
	user.Admin = d.GetBoolRef("admin")
	user.ProfileUpdatable = d.GetBoolRef("profile_updatable")
	user.DisableUIAccess = d.GetBoolRef("disable_ui_access")
	user.InternalPasswordDisabled = d.GetBoolRef("internal_password_disabled")
	user.Realm = d.GetStringRef("realm")
	user.Groups = d.GetSetRef("groups")

	return user
}

func marshalUser(user *artifactory.User, d *schema.ResourceData) {
	d.Set("name", user.Name)
	d.Set("email", user.Email)
	d.Set("admin", user.Admin)
	d.Set("profile_updatable", user.ProfileUpdatable)
	d.Set("disable_ui_access", user.DisableUIAccess)
	d.Set("realm", user.Realm)
	d.Set("internal_password_disabled", user.InternalPasswordDisabled)

	if user.Groups != nil {
		d.Set("groups", schema.NewSet(schema.HashString, CastToInterfaceArr(*user.Groups)))
	}
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	user := unmarshalUser(d)
	user.Password = artifactory.String(generatePassword())

	_, err := c.Security.CreateOrReplaceUser(context.Background(), *user.Name, user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resourceUserRead(d, m)
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	user, resp, err := c.Security.GetUser(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	marshalUser(user, d)
	return nil
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	user := unmarshalUser(d)
	_, err := c.Security.UpdateUser(context.Background(), d.Id(), user)
	if err != nil {
		return err
	}

	d.SetId(*user.Name)
	return resourceUserRead(d, m)
}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	user := unmarshalUser(d)
	_, resp, err := c.Security.DeleteUser(context.Background(), *user.Name)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

// generatePassword used as default func to generate user passwords
func generatePassword() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, randomPasswordLength)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}
