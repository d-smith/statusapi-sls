package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	auth0 "github.com/auth0-community/go-auth0"
	"os"
	"gopkg.in/square/go-jose.v2"
	"net/http"
	"fmt"
	"log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
)

var (
	tenantTable = os.Getenv("TENANT_TABLE")
	sess = session.New()
	svc = dynamodb.New(sess)
)

func allApiResources(methodArn string) []string {

	//API resource arns look like arn:aws:execute-api:<region>:<account>:<restapi id>/<stage>/<method>/uri

	arnParts := strings.Split(methodArn, ":")
	if len(arnParts) != 6 {
		log.Println("WARNING: expected 5 parts for method ARN", methodArn)
		return []string{methodArn}
	}

	arnBase := fmt.Sprintf("%s:%s:%s:%s:%s:", arnParts[0],arnParts[1],arnParts[2],arnParts[3],arnParts[4])
	restPart := arnParts[5]

	//Split the restPart on / - first entry will be the restapi id
	uriParts := strings.Split(restPart, "/")
	restapiId := uriParts[0]

	return []string {
		fmt.Sprintf("%s%s/*/POST/status/api/v1/events", arnBase,restapiId),
		fmt.Sprintf("%s%s/*/GET/status/api/v1/instances", arnBase, restapiId),
		fmt.Sprintf("%s%s/*/GET/status/api/v1/instances/*", arnBase, restapiId),
		fmt.Sprintf("%s%s/*/GET/status/api/v1/models", arnBase, restapiId),
		fmt.Sprintf("%s%s/*/POST/status/api/v1/models", arnBase, restapiId),
		fmt.Sprintf("%s%s/*/GET/status/api/v1/models/*", arnBase, restapiId),
		fmt.Sprintf("%s%s/*/PUT/status/api/v1/models/*", arnBase, restapiId),
	}

}

func getKeyForTenant(tenant string, ddbSvc *dynamodb.DynamoDB) (string, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"tenant": {
				S: aws.String(tenant),
			},
		},
		TableName: aws.String(tenantTable),
	}

	item,err := ddbSvc.GetItem(input)
	if err != nil {
		return "", err
	}

	if item == nil || item.Item == nil || item.Item["key"] == nil {
		return "", nil
	}

	return *item.Item["key"].S, nil
}



// Help function to generate an IAM policy
func generatePolicy(principalId, apiKey, effect, resource, tenant string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	allResources := allApiResources(resource)

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: allResources,
				},
			},
		}
	}

	// Optional output with custom properties of the String, Number or Boolean type.
	authResponse.Context = map[string]interface{} {
		"tenant":  tenant,
	}

	authResponse.UsageIdentifierKey = apiKey

	log.Printf("Authorizer response: %+v\n", authResponse)

	return authResponse
}

func handleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := event.AuthorizationToken
	fmt.Println("Request with token", token)


	JWKS_URI := "https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json"
	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: JWKS_URI})
	aud := os.Getenv("AUTH0_AUDIENCE")
	audience := []string{aud}

	var AUTH0_API_ISSUER = "https://" + os.Getenv("AUTH0_DOMAIN") + "/"
	configuration := auth0.NewConfiguration(client, audience, AUTH0_API_ISSUER, jose.RS256)
	validator := auth0.NewValidator(configuration)

	// Need to gin up a request to use the auth0 library
	fakeRequest,_ := http.NewRequest("GET","/",nil)
	fakeRequest.Header.Add("Authorization", token)
	jot, err := validator.ValidateRequest(fakeRequest)
	if err != nil {
		log.Println("WARNING", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Invalid token")
	}

	claims := map[string]interface{}{}

	err = validator.Claims(fakeRequest, jot, &claims)
	if err != nil {
		log.Println("Error looking at claims", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Invalid token")
	}

	tenant,ok := claims["https://status.aps-dev.net/tenant"].(string)
	if !ok ||tenant == "" {
		log.Println("Unable to extract tenant from token")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	log.Println("tenant from claim", tenant)

	apiKey, err := getKeyForTenant(tenant, svc)
	if err != nil || apiKey == "" {
		log.Println("No API key in tenant table for tenant", tenant)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	log.Println("key lookup ok", apiKey)

	name, ok := claims["name"].(string)
	if !ok || name == "" {
		log.Println("Unable to extract name (principal) from token")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	return generatePolicy(name, apiKey, "Allow", event.MethodArn, tenant), nil

}

func main() {
	lambda.Start(handleRequest)
}