package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	authWrapper "github.com/manicminer/hamilton-autorest/auth"
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/environments"
)

var (
	tenantId     = os.Getenv("TENANT_ID")
	clientId     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	environment  = os.Getenv("AZURE_ENVIRONMENT")
)

func TestAuthorizerWrapper(t *testing.T) {
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

	authorizer, err := authWrapper.NewAuthorizerWrapper(autorest.NewBearerAuthorizer(spt))
	if err != nil {
		t.Fatalf("NewAutorestAuthorizerWrapper(): %v", err)
	}
	if authorizer == nil {
		t.Fatal("auth is nil, expected Authorizer")
	}

	token, err := authorizer.Token()
	if err != nil {
		t.Fatalf("authorizer.Token(): %v", err)
	}
	if token == nil {
		t.Fatal("token was nil")
	}
	if token.AccessToken == "" {
		t.Fatal("token.AccessToken was empty")
	}
}

func TestAuthorizerWithAuthorization(t *testing.T) {
	ctx := context.Background()

	env, err := environments.EnvironmentFromString(environment)
	if err != nil {
		t.Fatal(err)
	}

	authorizer, err := setupAuthorizer(ctx, env)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := &authWrapper.Authorizer{Authorizer: authorizer}
	if err = testWithAuthorization(wrapper, env.MsGraph.Resource()); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorizerBearerAuthorizerCallback(t *testing.T) {
	ctx := context.Background()

	env, err := environments.EnvironmentFromString(environment)
	if err != nil {
		t.Fatal(err)
	}

	authorizer, err := setupAuthorizer(ctx, env)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := &authWrapper.Authorizer{Authorizer: authorizer}

	callback := wrapper.BearerAuthorizerCallback()
	if err = testWithAuthorization(callback, "https://contoso.vault.azure.net/secrets"); err != nil {
		t.Fatal(err)
	}
}

type preparer struct{}

func (preparer) Prepare(r *http.Request) (*http.Request, error) {
	return r, nil
}

func setupAuthorizer(ctx context.Context, env environments.Environment) (auth.Authorizer, error) {
	authConfig := &auth.Config{
		Environment:            env,
		TenantID:               tenantId,
		ClientID:               clientId,
		ClientSecret:           clientSecret,
		EnableClientSecretAuth: true,
	}

	authorizer, err := authConfig.NewAuthorizer(ctx, env.MsGraph)
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

func testWithAuthorization(authorizer autorest.Authorizer, resource string) error {
	u, err := url.Parse(resource)
	if err != nil {
		return err
	}

	r := &http.Request{
		URL: u,
	}

	r, err = authorizer.WithAuthorization()(preparer{}).Prepare(r)
	if err != nil {
		return err
	}

	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return errors.New("WithAuthorization(): Authorization header has no bearer token")
	}

	return nil
}
