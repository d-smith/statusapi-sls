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
curl -H "x-api-key: UKslut46mV9MN21TlYGzf2SB0cCyyJVp411yCXka" -XPOST -d '{"name":"model1", "steps":["s1", "s2", "s3"]}' https://e5vii1knla.execute-api.us-east-1.amazonaws.com/dev/models


curl -H "x-api-key: UKslut46mV9MN21TlYGzf2SB0cCyyJVp411yCXka" -XPOST -d '{"txn_id":"1a","event_id":"1","step":"s1","step_state":"completed"}' https://e5vii1knla.execute-api.us-east-1.amazonaws.com/dev/events

curl -H "x-api-key: UKslut46mV9MN21TlYGzf2SB0cCyyJVp411yCXka" -XPOST -d '{"txn_id":"1a","event_id":"2","step":"s2","step_state":"completed"}' https://e5vii1knla.execute-api.us-east-1.amazonaws.com/dev/events
 
curl -H "x-api-key: UKslut46mV9MN21TlYGzf2SB0cCyyJVp411yCXka"  'https://e5vii1knla.execute-api.us-east-1.amazonaws.com/dev/instances/1a?model=model1'
```

