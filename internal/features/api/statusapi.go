package api

import (
	"bytes"
	"errors"
	"fmt"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func postEventsForModel(apiKey, apiEndpoint string) error {
	eventPostEndpoint := fmt.Sprintf("https://%s/dev/status/api/v1/events", apiEndpoint)
	log.Println("send to", eventPostEndpoint)
	txnId := fmt.Sprintf("txn-%d", rand.Int())
	for i := 0; i < 2; i += 1 {
		payload := fmt.Sprintf(`{"txn_id":"%s","event_id":"%d","step":"s%d","step_state":"completed"}`, txnId, i, i+1)
		log.Println("sending payload", payload)

		req, err := http.NewRequest("POST", eventPostEndpoint, bytes.NewBuffer([]byte(payload)))
		if !assert.Nil(T, err) {
			return err
		}
		req.Header.Add("x-api-key", apiKey)

		client := &http.Client{}
		log.Println("post event")
		resp, err := client.Do(req)
		if !assert.Nil(T, err) {
			log.Printf("error on event request: %s", err.Error())
			return err
		}

		if !assert.Equal(T, http.StatusOK, resp.StatusCode) {
			return errors.New(fmt.Sprintf("Unexcepted status code %d", resp.StatusCode))
		}
	}

	return nil
}

func init() {

	rand.Seed(time.Now().UnixNano())
	var (
		apiKey      = os.Getenv("APIKEY")
		apiEndpoint = os.Getenv("API_ENDPOINT")
		testBase    = fmt.Sprintf("x%d", rand.Int())
	)

	Given(`^a milestone model$`, func() {
		//curl -H "x-api-key: XXXX" -XPOST -d '{"name":"model1", "steps":["s1", "s2", "s3"]}' https://ENDPOINT/dev/status/api/v1/models
		modelPostUrl := fmt.Sprintf("https://%s/dev/status/api/v1/models", apiEndpoint)
		log.Printf("request with api key %s going to %s", apiKey, modelPostUrl)
		payload := fmt.Sprintf(`{"name":"model%s", "steps":["s1", "s2", "s3"]}`, testBase)
		req, err := http.NewRequest("POST", modelPostUrl, bytes.NewBuffer([]byte(payload)))
		if !assert.Nil(T, err) {
			return
		}
		req.Header.Add("x-api-key", apiKey)

		client := &http.Client{}
		log.Println("make test request")
		resp, err := client.Do(req)
		if !assert.Nil(T, err) {
			log.Printf("error on test request: %s", err.Error())
			return
		}

		if !assert.Equal(T, http.StatusOK, resp.StatusCode) {
			return
		}
		log.Printf(fmt.Sprintf("Milestone model model%s created", testBase))
	})

	And(`^some correlated events for the model$`, func() {
		//curl -H "x-api-key: XXXX" -XPOST -d '{"txn_id":"1a","event_id":"1","step":"s1","step_state":"completed"}' https://ENDPOINT/dev/status/api/v1/events
		postEventsForModel(apiKey, apiEndpoint)
	})

	When(`^I retrieve the model state for the correlated events$`, func() {
		//T.Skip() // pending
	})

	Then(`^the state of the model reflects the events$`, func() {
		//T.Skip() // pending
	})

}
