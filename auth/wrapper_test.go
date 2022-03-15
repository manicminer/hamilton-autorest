package auth_test

import (
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/environments"
)

var (
	tenantId     = os.Getenv("TENANT_ID")
	clientId     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	environment  = os.Getenv("AZURE_ENVIRONMENT")
)

func TestAutorestAuthorizerWrapper(t *testing.T) {
	env, err := environments.EnvironmentFromString(environment)
	if err != nil {
		t.Fatal(err)
	}

	// adal.ServicePrincipalToken.refreshInternal() doesn't support v2 tokens
	version := "1.0"
	oauthConfig, err := adal.NewOAuthConfigWithAPIVersion(string(env.AzureADEndpoint), tenantId, &version)
	if err != nil {
		t.Fatalf("adal.NewOAuthConfig(): %v", err)
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientId, clientSecret, string(env.MsGraph.Endpoint))
	if err != nil {
		t.Fatalf("adal.NewServicePrincipalToken(): %v", err)
	}

	auth, err := auth.NewAutorestAuthorizerWrapper(autorest.NewBearerAuthorizer(spt))
	if err != nil {
		t.Fatalf("NewAutorestAuthorizerWrapper(): %v", err)
	}
	if auth == nil {
		t.Fatal("auth is nil, expected Authorizer")
	}

	token, err := auth.Token()
	if err != nil {
		t.Fatalf("auth.Token(): %v", err)
	}
	if token == nil {
		t.Fatal("token was nil")
	}
	if token.AccessToken == "" {
		t.Fatal("token.AccessToken was empty")
	}
}
