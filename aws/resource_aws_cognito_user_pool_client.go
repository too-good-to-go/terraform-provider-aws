package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCognitoUserPoolClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolClientCreate,
		Read:   resourceAwsCognitoUserPoolClientRead,
		Update: resourceAwsCognitoUserPoolClientUpdate,
		Delete: resourceAwsCognitoUserPoolClientDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsCognitoUserPoolClientImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPoolClient.html
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`[\w\s+=,.@-]+`),
						"must satisfy regular expression pattern: `[\\w\\s+=,.@-]+`"),
				),
			},

			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"generate_secret": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"explicit_auth_flows": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.ExplicitAuthFlowsType_Values(), false),
				},
			},

			"read_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"write_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"refresh_token_validity": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(0, 3650),
			},
			"access_token_validity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"id_token_validity": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"allowed_oauth_flows": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.OAuthFlowType_Values(), false),
				},
			},

			"allowed_oauth_flows_user_pool_client": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"allowed_oauth_scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// https://docs.aws.amazon.com/cognito/latest/developerguide/authorization-endpoint.html
					// System reserved scopes are openid, email, phone, profile, and aws.cognito.signin.user.admin.
					// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPoolClient.html#CognitoUserPools-CreateUserPoolClient-request-AllowedOAuthScopes
					// Constraints seem like to be designed for custom scopes which are not supported yet?
				},
			},

			"callback_urls": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},

			"default_redirect_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
						"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
				),
			},

			"logout_urls": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},

			"prevent_user_existence_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.PreventUserExistenceErrorTypes_Values(), false),
			},

			"supported_identity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 32),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},
			"token_validity_units": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeHours,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
						"id_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeHours,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
						"refresh_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeDays,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
					},
				},
			},
			"analytics_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"user_data_shared": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsCognitoUserPoolClientCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName: aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("generate_secret"); ok {
		params.GenerateSecret = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("explicit_auth_flows"); ok {
		params.ExplicitAuthFlows = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("refresh_token_validity"); ok {
		params.RefreshTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("access_token_validity"); ok {
		params.AccessTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("id_token_validity"); ok {
		params.IdTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allowed_oauth_flows"); ok {
		params.AllowedOAuthFlows = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("allowed_oauth_flows_user_pool_client"); ok {
		params.AllowedOAuthFlowsUserPoolClient = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("allowed_oauth_scopes"); ok {
		params.AllowedOAuthScopes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("callback_urls"); ok {
		params.CallbackURLs = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_redirect_uri"); ok {
		params.DefaultRedirectURI = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logout_urls"); ok {
		params.LogoutURLs = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("supported_identity_providers"); ok {
		params.SupportedIdentityProviders = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("analytics_configuration"); ok {
		params.AnalyticsConfiguration = expandAwsCognitoUserPoolClientAnalyticsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("token_validity_units"); ok {
		params.TokenValidityUnits = expandAwsCognitoUserPoolClientTokenValidityUnitsType(v.([]interface{}))
	}

	if v, ok := d.GetOk("prevent_user_existence_errors"); ok {
		params.PreventUserExistenceErrors = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Client: %s", params)

	resp, err := conn.CreateUserPoolClient(params)

	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool Client: %s", err)
	}

	d.SetId(aws.StringValue(resp.UserPoolClient.ClientId))

	return resourceAwsCognitoUserPoolClientRead(d, meta)
}

func resourceAwsCognitoUserPoolClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool Client: %s", params)

	resp, err := conn.DescribeUserPoolClient(params)

	if err != nil {
		if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Cognito User Pool Client %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	userPoolClient := resp.UserPoolClient
	d.Set("user_pool_id", userPoolClient.UserPoolId)
	d.Set("name", userPoolClient.ClientName)
	d.Set("explicit_auth_flows", flattenStringSet(userPoolClient.ExplicitAuthFlows))
	d.Set("read_attributes", flattenStringSet(userPoolClient.ReadAttributes))
	d.Set("write_attributes", flattenStringSet(userPoolClient.WriteAttributes))
	d.Set("refresh_token_validity", userPoolClient.RefreshTokenValidity)
	d.Set("access_token_validity", userPoolClient.AccessTokenValidity)
	d.Set("id_token_validity", userPoolClient.IdTokenValidity)
	d.Set("client_secret", userPoolClient.ClientSecret)
	d.Set("allowed_oauth_flows", flattenStringSet(userPoolClient.AllowedOAuthFlows))
	d.Set("allowed_oauth_flows_user_pool_client", userPoolClient.AllowedOAuthFlowsUserPoolClient)
	d.Set("allowed_oauth_scopes", flattenStringSet(userPoolClient.AllowedOAuthScopes))
	d.Set("callback_urls", flattenStringSet(userPoolClient.CallbackURLs))
	d.Set("default_redirect_uri", userPoolClient.DefaultRedirectURI)
	d.Set("logout_urls", flattenStringSet(userPoolClient.LogoutURLs))
	d.Set("prevent_user_existence_errors", userPoolClient.PreventUserExistenceErrors)
	d.Set("supported_identity_providers", flattenStringSet(userPoolClient.SupportedIdentityProviders))

	if err := d.Set("analytics_configuration", flattenAwsCognitoUserPoolClientAnalyticsConfig(userPoolClient.AnalyticsConfiguration)); err != nil {
		return fmt.Errorf("error setting analytics_configuration: %s", err)
	}

	if err := d.Set("token_validity_units", flattenAwsCognitoUserPoolClientTokenValidityUnitsType(userPoolClient.TokenValidityUnits)); err != nil {
		return fmt.Errorf("error setting token_validity_units: %w", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolClientUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.UpdateUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		params.ClientName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("explicit_auth_flows"); ok {
		params.ExplicitAuthFlows = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("refresh_token_validity"); ok {
		params.RefreshTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("access_token_validity"); ok {
		params.AccessTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("id_token_validity"); ok {
		params.IdTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allowed_oauth_flows"); ok {
		params.AllowedOAuthFlows = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("allowed_oauth_flows_user_pool_client"); ok {
		params.AllowedOAuthFlowsUserPoolClient = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("allowed_oauth_scopes"); ok {
		params.AllowedOAuthScopes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("callback_urls"); ok {
		params.CallbackURLs = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_redirect_uri"); ok {
		params.DefaultRedirectURI = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logout_urls"); ok {
		params.LogoutURLs = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("prevent_user_existence_errors"); ok {
		params.PreventUserExistenceErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_identity_providers"); ok {
		params.SupportedIdentityProviders = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("analytics_configuration"); ok {
		params.AnalyticsConfiguration = expandAwsCognitoUserPoolClientAnalyticsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("token_validity_units"); ok {
		params.TokenValidityUnits = expandAwsCognitoUserPoolClientTokenValidityUnitsType(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating Cognito User Pool Client: %s", params)

	_, err := conn.UpdateUserPoolClient(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito User Pool Client: %s", err)
	}

	return resourceAwsCognitoUserPoolClientRead(d, meta)
}

func resourceAwsCognitoUserPoolClientDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool Client: %s", params)

	_, err := conn.DeleteUserPoolClient(params)

	if err != nil {
		return fmt.Errorf("Error deleting Cognito User Pool Client: %s", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolClientImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 || len(d.Id()) < 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'user-pool-id/client-id'", d.Id())
	}
	userPoolId := strings.Split(d.Id(), "/")[0]
	clientId := strings.Split(d.Id(), "/")[1]
	d.SetId(clientId)
	d.Set("user_pool_id", userPoolId)
	log.Printf("[DEBUG] Importing client %s for user pool %s", clientId, userPoolId)

	return []*schema.ResourceData{d}, nil
}

func expandAwsCognitoUserPoolClientAnalyticsConfig(l []interface{}) *cognitoidentityprovider.AnalyticsConfigurationType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	analyticsConfig := &cognitoidentityprovider.AnalyticsConfigurationType{
		ApplicationId: aws.String(m["application_id"].(string)),
		ExternalId:    aws.String(m["external_id"].(string)),
		RoleArn:       aws.String(m["role_arn"].(string)),
	}

	if v, ok := m["user_data_shared"]; ok {
		analyticsConfig.UserDataShared = aws.Bool(v.(bool))
	}

	return analyticsConfig
}

func flattenAwsCognitoUserPoolClientAnalyticsConfig(analyticsConfig *cognitoidentityprovider.AnalyticsConfigurationType) []interface{} {
	if analyticsConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"application_id":   aws.StringValue(analyticsConfig.ApplicationId),
		"external_id":      aws.StringValue(analyticsConfig.ExternalId),
		"role_arn":         aws.StringValue(analyticsConfig.RoleArn),
		"user_data_shared": aws.BoolValue(analyticsConfig.UserDataShared),
	}

	return []interface{}{m}
}

func expandAwsCognitoUserPoolClientTokenValidityUnitsType(l []interface{}) *cognitoidentityprovider.TokenValidityUnitsType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	tokenValidityConfig := &cognitoidentityprovider.TokenValidityUnitsType{}

	if v, ok := m["access_token"]; ok {
		tokenValidityConfig.AccessToken = aws.String(v.(string))
	}

	if v, ok := m["id_token"]; ok {
		tokenValidityConfig.IdToken = aws.String(v.(string))
	}

	if v, ok := m["refresh_token"]; ok {
		tokenValidityConfig.RefreshToken = aws.String(v.(string))
	}

	return tokenValidityConfig
}

func flattenAwsCognitoUserPoolClientTokenValidityUnitsType(tokenValidityConfig *cognitoidentityprovider.TokenValidityUnitsType) []interface{} {
	if tokenValidityConfig == nil {
		return nil
	}

	//tokenValidityConfig is never nil and if everything is empty it causes diffs
	if tokenValidityConfig.IdToken == nil && tokenValidityConfig.AccessToken == nil && tokenValidityConfig.RefreshToken == nil {
		return nil
	}

	m := map[string]interface{}{}

	if tokenValidityConfig.IdToken != nil {
		m["id_token"] = aws.StringValue(tokenValidityConfig.IdToken)
	}

	if tokenValidityConfig.AccessToken != nil {
		m["access_token"] = aws.StringValue(tokenValidityConfig.AccessToken)
	}

	if tokenValidityConfig.RefreshToken != nil {
		m["refresh_token"] = aws.StringValue(tokenValidityConfig.RefreshToken)
	}

	return []interface{}{m}
}
