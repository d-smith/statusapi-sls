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
)

// Help function to generate an IAM policy
func generatePolicy(principalId, effect, resource, tenent string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	// Optional output with custom properties of the String, Number or Boolean type.
	authResponse.Context = map[string]interface{} {
		"tenent":  tenent,
	}
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
		fmt.Println("WARNING", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: Invalid token")
	}

	claims := map[string]interface{}{}

	err = validator.Claims(fakeRequest, jot, &claims)
	if err != nil {
		log.Println("Error looking at claims", err.Error())
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: Invalid token")
	}

	tenent := claims["https://status.aps-dev.net/tenent"]
	if tenent == "" {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	name := claims["name"]

	return generatePolicy(name.(string), "Allow", event.MethodArn, tenent.(string)), nil

}

func main() {
	lambda.Start(handleRequest)
}