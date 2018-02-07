package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func postEventsForModel(apiKey, apiEndpoint string) (string, error) {
	eventPostEndpoint := fmt.Sprintf("https://%s/dev/status/api/v1/events", apiEndpoint)
	log.Println("send to", eventPostEndpoint)
	txnId := fmt.Sprintf("txn-%d", rand.Int())
	for i := 0; i < 2; i += 1 {
		payload := fmt.Sprintf(`{"txn_id":"%s","event_id":"%d","step":"s%d","step_state":"completed"}`, txnId, i, i+1)
		log.Println("sending payload", payload)

		req, err := http.NewRequest("POST", eventPostEndpoint, bytes.NewBuffer([]byte(payload)))
		if !assert.Nil(T, err) {
			return "", err
		}
		req.Header.Add("x-api-key", apiKey)

		client := &http.Client{}
		log.Println("post event")
		resp, err := client.Do(req)
		if !assert.Nil(T, err) {
			log.Printf("error on event request: %s", err.Error())
			return "", err
		}

		if !assert.Equal(T, http.StatusOK, resp.StatusCode) {
			return "", errors.New(fmt.Sprintf("Unexcepted status code %d", resp.StatusCode))
		}
	}

	return txnId, nil
}

func init() {

	rand.Seed(time.Now().UnixNano())
	var (
		apiKey      = os.Getenv("APIKEY")
		apiEndpoint = os.Getenv("API_ENDPOINT")
		testBase    = fmt.Sprintf("x%d", rand.Int())
		txnId       = ""
		modelState  = ""
	)

	Given(`^a milestone model$`, func() {
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
		var err error
		txnId, err = postEventsForModel(apiKey, apiEndpoint)
		if !assert.Nil(T, err) {
			log.Printf("error on posting events: %s", err.Error())
			return
		}
	})

	When(`^I retrieve the model state for the correlated events$`, func() {
		//curl -H "x-api-key: XXXX"  'https://ENDPOINT/dev/status/api/v1/instances/1a?model=model1'
		requestUrl := fmt.Sprintf("https://%s/dev/status/api/v1/instances/%s?model=model%s", apiEndpoint, txnId, testBase)
		log.Println("get", requestUrl)

		req, err := http.NewRequest("GET", requestUrl, nil)
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

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		modelState = string(body)
	})

	Then(`^the state of the model reflects the events$`, func() {

		type ModelState struct {
			Step  string `json:"step"`
			State string `json:"step_state"`
		}

		//T.Skip() // pending
		log.Println(modelState)

		var states []ModelState
		err := json.Unmarshal([]byte(modelState), &states)
		if !assert.Nil(T, err) {
			log.Printf("error parsing response: %s", err.Error())
			return
		}

		assert.Equal(T, 3, len(states), "Unexpected number of states parsed from response")
		stateMap := make(map[string]string)
		for _, s := range states {
			stateMap[s.Step] = s.State
		}

		assert.Equal(T, "completed", stateMap["s1"])
		assert.Equal(T, "completed", stateMap["s2"])
		assert.Equal(T, "", stateMap["s3"])

	})

}
