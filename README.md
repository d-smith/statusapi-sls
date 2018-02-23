# status api

Serverless framework version of a simple status api written using golang.

This template is useful for bootstrapping a new project:

```
serverless create -u https://github.com/serverless/serverless-golang/ -p myservice
```

Use the make file to build the binaries needed to deploy. To deploy:

```
serverless deploy --aws-profile <profile>
```

Example event payload:

```
{"txn_id":"1a","event_id":"1","step":"Order Received","step_state":"active"}
```

Example model definition:

```
{"name":"model1", "steps":["s1", "s2", "s3"]}
```

Simple scenario - define a model, post some events, retrieve view
of model based on instance state

```
curl -H "x-api-key: XXXX" -XPOST -d '{"name":"model1", "steps":["s1", "s2", "s3"]}' https://ENDPOINT/dev/status/api/v1/models

curl -H "x-api-key: XXXX" -XPOST -d '{"txn_id":"1a","event_id":"1","step":"s1","step_state":"completed"}' https://ENDPOINT/dev/status/api/v1/events

curl -H "x-api-key: XXXX" -XPOST -d '{"txn_id":"1a","event_id":"2","step":"s2","step_state":"completed"}' https://ENDPOINT/dev/status/api/v1/events
 
curl -H "x-api-key: XXXX"  'https://ENDPOINT/dev/status/api/v1/instances/1a?model=model1'
```


## Multi-tenant support

Note: this is still being fleshed out

The basic premise is:

1. Set up users in an Auto0 domain, and include a tenant attribute in
their user_metadata

2. Use a rule to inject the tenant as a claim into the identity token
produced by Auth0

3. In a custom authorizer, validate the token, then look up the api
key for the tenant.

Currently there does not appear to be support for providing the proper
settings at the rest api definition in cloud front, so post
install the cli has to be used to set api key source from
header to authorizer.


<pre>
# Grab the api id, which is the first component of the expoint
# url, or use the cli
aws apigateway get-rest-apis

# Update the settings using the CLI
aws apigateway update-rest-api --rest-api-id o304m2z79a --patch-operations op=replace,path=/apiKeySource,value=AUTHORIZER
</pre>

Note that the API must then be redeployed via sls deploy for the API
settings update to take effect. So the deploy process is:

* Deploy
* Configure gateway settings
* Redeploy
* Seed DDB with tenant keys
