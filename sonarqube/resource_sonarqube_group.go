package sonarqube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetGroup for unmarshalling response body from getting group details
type GetGroup struct {
	Paging Paging  `json:"paging"`
	Groups []Group `json:"groups"`
}

// CreateGroupResponse for unmarshalling response body of group creation
type CreateGroupResponse struct {
	Group Group `json:"group"`
}

// Group struct
type Group struct {
	ID           string   `json:"id,omitempty"`
	Organization string   `json:"organization,omitempty"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	MembersCount int      `json:"membersCount,omitempty"`
	IsDefault    bool     `json:"default,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
}

// Returns the resource represented by this file.
func resourceSonarqubeGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSonarqubeGroupCreate,
		Read:   resourceSonarqubeGroupRead,
		Update: resourceSonarqubeGroupUpdate,
		Delete: resourceSonarqubeGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSonarqubeGroupImport,
		},

		// Define the fields of this schema.
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSonarqubeGroupCreate(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = strings.TrimSuffix(sonarQubeURL.Path, "/") + "/api/user_groups/create"
	sonarQubeURL.RawQuery = url.Values{
		"name":        []string{d.Get("name").(string)},
		"description": []string{d.Get("description").(string)},
	}.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusOK,
		"resourceSonarqubeGroupCreate",
	)
	if err != nil {
		return fmt.Errorf("error creating Sonarqube group: %+v", err)
	}
	defer resp.Body.Close()

	// Decode response into struct
	groupResponse := CreateGroupResponse{}
	err = json.NewDecoder(resp.Body).Decode(&groupResponse)
	if err != nil {
		return fmt.Errorf("resourceSonarqubeGroupRead: Failed to decode json into struct: %+v", err)
	}
	d.SetId(groupResponse.Group.ID)

	return resourceSonarqubeGroupRead(d, m)
}

func resourceSonarqubeGroupRead(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = strings.TrimSuffix(sonarQubeURL.Path, "/") + "/api/user_groups/search"
	sonarQubeURL.RawQuery = url.Values{
		"ps": []string{"500"},
		"q":  []string{d.Get("name").(string)},
	}.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"GET",
		sonarQubeURL.String(),
		http.StatusOK,
		"resourceSonarqubeGroupRead",
	)
	if err != nil {
		return fmt.Errorf("error reading Sonarqube group: %+v", err)
	}
	defer resp.Body.Close()

	readSuccess := false
	// Decode response into struct
	groupReadResponse := GetGroup{}
	err = json.NewDecoder(resp.Body).Decode(&groupReadResponse)
	if err != nil {
		return fmt.Errorf("resourceSonarqubeGroupRead: Failed to decode json into struct: %+v", err)
	}
	// Loop over all groups to see if the group we need exists.
	for _, value := range groupReadResponse.Groups {
		if d.Id() == value.ID {
			// If it does, set the values of that group
			d.SetId(value.ID)
			d.Set("name", value.Name)
			d.Set("description", value.Description)
			readSuccess = true
		}
	}

	if !readSuccess {
		// Group not found
		d.SetId("")
	}

	return nil
}

func resourceSonarqubeGroupUpdate(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = strings.TrimSuffix(sonarQubeURL.Path, "/") + "/api/user_groups/update"

	rawQuery := url.Values{
		"id": []string{d.Id()},
	}

	if _, ok := d.GetOk("description"); ok {
		rawQuery.Add("description", d.Get("description").(string))
	} else {
		rawQuery.Add("description", "")
	}

	sonarQubeURL.RawQuery = rawQuery.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusOK,
		"resourceSonarqubeGroupUpdate",
	)
	if err != nil {
		return fmt.Errorf("error updating Sonarqube group: %+v", err)
	}
	defer resp.Body.Close()

	return resourceSonarqubeGroupRead(d, m)
}

func resourceSonarqubeGroupDelete(d *schema.ResourceData, m interface{}) error {
	sonarQubeURL := m.(*ProviderConfiguration).sonarQubeURL
	sonarQubeURL.Path = strings.TrimSuffix(sonarQubeURL.Path, "/") + "/api/user_groups/delete"

	sonarQubeURL.RawQuery = url.Values{
		"id": []string{d.Id()},
	}.Encode()

	resp, err := httpRequestHelper(
		m.(*ProviderConfiguration).httpClient,
		"POST",
		sonarQubeURL.String(),
		http.StatusNoContent,
		"resourceSonarqubeGroupDelete",
	)
	if err != nil {
		return fmt.Errorf("error deleting Sonarqube group: %+v", err)
	}
	defer resp.Body.Close()

	return nil
}

func resourceSonarqubeGroupImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSonarqubeGroupRead(d, m); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
