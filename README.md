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
{"correlation_id":"1a","event_id":"1","model_ids":["model_1"],"state":"Order Received"}
```

Example model definition:

```
{"name":"model1", "states":["s1", "s2", "s3"]}
```
