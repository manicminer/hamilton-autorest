## Autorest Wrappers for Hamilton SDK

This module contains wrappers for [Azure/go-autorest](https://github.com/Azure/go-autorest) for the [Hamilton SDK](https://github.com/manicminer/hamilton).

It is published as a separate module to avoid pulling a dependency on go-autorest in the main Hamilton library.

## Example: Consuming an autorest.Authorizer

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	authWrapper "github.com/manicminer/hamilton-autorest/auth"
	"github.com/manicminer/hamilton/environments"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
)

var (
	tenantId     = os.Getenv("TENANT_ID")
	clientId     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
)

func main() {
	ctx := context.Background()
	environment := environments.Global

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

	wrapper, err := authWrapper.NewAuthorizerWrapper(autorest.NewBearerAuthorizer(spt))
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
```

## Example: Dispensing an autorest.Authorizer

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-10-01/resources"
	authWrapper "github.com/manicminer/hamilton-autorest/auth"
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/environments"
)

var (
	tenantId       = os.Getenv("TENANT_ID")
	subscriptionId = os.Getenv("SUBSCRIPTION_ID")
	clientId       = os.Getenv("CLIENT_ID")
	clientSecret   = os.Getenv("CLIENT_SECRET")
)

func main() {
	ctx := context.Background()
	environment := environments.Global

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

	client := resources.NewClientWithBaseURI(string(environment.ResourceManager.Endpoint), subscriptionId)
	client.Client.Authorizer = &authWrapper.Authorizer{Authorizer: authorizer}

	for resp, err := client.ListComplete(ctx, "", "", nil); resp.NotDone(); err = resp.NextWithContext(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%s)\n", *resp.Value().Name, *resp.Value().Location)
	}
}
```
