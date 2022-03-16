package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"

	authWrapper "github.com/manicminer/hamilton-autorest/auth"
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/environments"
)

var (
	tenantId     = os.Getenv("TENANT_ID")
	clientId     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
)

func autorestAuthorizer(environment environments.Environment) autorest.Authorizer {
	oauth, err := adal.NewOAuthConfig(string(environment.AzureADEndpoint), tenantId)
	if err != nil {
		log.Fatal(err)
	}
	if oauth == nil {
		log.Fatalf("OAuthConfig was nil for tenant %s", tenantId)
	}

	spt, err := adal.NewServicePrincipalToken(*oauth, clientId, clientSecret, environment.MsGraph.Resource())
	if err != nil {
		log.Fatal(err)
	}

	return autorest.NewBearerAuthorizer(spt)
}

func hamiltonAuthorizer(ctx context.Context, environment environments.Environment) auth.Authorizer {
	authConfig := &auth.Config{
		Environment:            environment,
		TenantID:               tenantId,
		ClientID:               clientId,
		ClientSecret:           clientSecret,
		EnableClientSecretAuth: true,
	}

	authorizer, err := authConfig.NewAuthorizer(ctx, environment.ResourceManager)
	if err != nil {
		log.Fatal(err)
	}

	return authorizer
}

func consumeExample(ctx context.Context, environment environments.Environment) {
	wrapper, err := authWrapper.NewAuthorizerWrapper(autorestAuthorizer(environment))
	if err != nil {
		log.Fatal(err)
	}

	client := msgraph.NewUsersClient(tenantId)
	client.BaseClient.Authorizer = wrapper

	users, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		log.Fatal(err)
	}
	if users == nil {
		log.Fatalln("bad API response, nil result received")
	}

	for _, user := range *users {
		fmt.Printf("%s: %s <%s>\n", *user.ID, *user.DisplayName, *user.UserPrincipalName)
	}
}

func dispenseExample(ctx context.Context, environment environments.Environment) {

	client := struct {
		autorest.Client
	}{}

	// client := resources.NewClientWithBaseURI(string(environment.ResourceManager.Endpoint), os.Getenv("SUBSCRIPTION_ID"))

	client.Client.Authorizer = &authWrapper.Authorizer{Authorizer: hamiltonAuthorizer(ctx, environment)}
}

func main() {
	ctx := context.Background()
	environment := environments.Global

	consumeExample(ctx, environment)
	dispenseExample(ctx, environment)
}
