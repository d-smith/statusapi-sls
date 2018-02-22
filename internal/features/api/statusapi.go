package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d-smith/statusapi-sls/model"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func postEventsForModel(apiKey, apiEndpoint, idToken string) (string, error) {
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
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

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

func postNewEventForModel(apiKey, apiEndpoint, idToken, txnId string) error {
	eventPostEndpoint := fmt.Sprintf("https://%s/dev/status/api/v1/events", apiEndpoint)
	log.Println("send to", eventPostEndpoint)
	payload := fmt.Sprintf(`{"txn_id":"%s","event_id":"100","step":"ess3","step_state":"completed"}`, txnId)
	log.Println("sending payload", payload)

	req, err := http.NewRequest("POST", eventPostEndpoint, bytes.NewBuffer([]byte(payload)))
	if !assert.Nil(T, err) {
		return err
	}
	req.Header.Add("x-api-key", apiKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

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

	return nil
}

func retrieveModelState(apiKey, apiEndpoint, idToken, txnId, testBase string) (string, error) {
	//curl -H "x-api-key: XXXX"  'https://ENDPOINT/dev/status/api/v1/instances/1a?model=model1'
	requestUrl := fmt.Sprintf("https://%s/dev/status/api/v1/instances/%s?model=model%s", apiEndpoint, txnId, testBase)
	log.Println("get", requestUrl)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("x-api-key", apiKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

	client := &http.Client{}
	log.Println("make test request")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error on test request: %s", err.Error())
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("http call status not ok")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func init() {

	rand.Seed(time.Now().UnixNano())
	var (
		apiKey      = os.Getenv("APIKEY")
		apiEndpoint = os.Getenv("API_ENDPOINT")
		idToken = os.Getenv("STATUS_ID_TOKEN")
		testBase    = fmt.Sprintf("x%d", rand.Int())
		txnId       = ""
		modelState  = ""
	)

	if apiKey == "" || apiEndpoint == "" {
		log.Println("Must set both APIKEY and API_ENDPOINT environment variables to run gucumber tests")
		os.Exit(1)
	}

	Given(`^a milestone model$`, func() {
		modelPostUrl := fmt.Sprintf("https://%s/dev/status/api/v1/models", apiEndpoint)
		log.Printf("request with api key %s going to %s", apiKey, modelPostUrl)
		payload := fmt.Sprintf(`{"name":"model%s", "steps":["s1", "s2", "s3"]}`, testBase)
		log.Println("posting", payload)
		req, err := http.NewRequest("POST", modelPostUrl, bytes.NewBuffer([]byte(payload)))
		if !assert.Nil(T, err) {
			return
		}
		req.Header.Add("x-api-key", apiKey)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

		log.Println(req.Header)

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
		txnId, err = postEventsForModel(apiKey, apiEndpoint, idToken)
		if !assert.Nil(T, err) {
			log.Printf("error on posting events: %s", err.Error())
			return
		}
	})

	When(`^I retrieve the model state for the correlated events$`, func() {
		var err error
		modelState, err = retrieveModelState(apiKey, apiEndpoint, idToken, txnId, testBase)
		assert.Nil(T, err)
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

	Given(`^a milestone model and correlated events$`, func() {
		//Created above
	})

	When(`^I update the model$`, func() {
		modelPutUrl := fmt.Sprintf("https://%s/dev/status/api/v1/models/model%s", apiEndpoint, testBase)
		payload := fmt.Sprintf(`{"name":"model%s", "steps":["s1", "s2", "ess3"]}`, testBase)
		req, err := http.NewRequest("PUT", modelPutUrl, bytes.NewBuffer([]byte(payload)))
		if !assert.Nil(T, err) {
			return
		}
		req.Header.Add("x-api-key", apiKey)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

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
		log.Printf(fmt.Sprintf("Milestone model model%s updated", testBase))
	})

	Then(`^the model state reflects the update$`, func() {
		type ModelState struct {
			Step  string `json:"step"`
			State string `json:"step_state"`
		}

		log.Println(modelState)

		newModelState, err := retrieveModelState(apiKey, apiEndpoint, idToken, txnId, testBase)
		if !assert.Nil(T, err) {
			return
		}

		var states []ModelState
		err = json.Unmarshal([]byte(newModelState), &states)
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
		assert.Equal(T, "", stateMap["ess3"])
	})

	And(`^the model update is durable$`, func() {
		//curl -H "x-api-key: XXXX"  'https://ENDPOINT/dev/status/api/v1/instances/1a?model=model1'
		requestUrl := fmt.Sprintf("https://%s/dev/status/api/v1/models/model%s", apiEndpoint, testBase)
		log.Println("get", requestUrl)

		req, err := http.NewRequest("GET", requestUrl, nil)
		if !assert.Nil(T, err) {
			return
		}
		req.Header.Add("x-api-key", apiKey)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

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
		modelUpdate := string(body)
		log.Println("model update: ", modelUpdate)

		var model model.Model
		err = json.Unmarshal([]byte(modelUpdate), &model)
		if assert.Nil(T, err) {
			assert.Contains(T, model.Steps, "s1")
			assert.Contains(T, model.Steps, "s2")
			assert.Contains(T, model.Steps, "ess3")
		}
	})

	And(`^the model reflects future events$`, func() {
		err := postNewEventForModel(apiKey, apiEndpoint, idToken, txnId)
		if !assert.Nil(T, err) {
			return
		}

		type ModelState struct {
			Step  string `json:"step"`
			State string `json:"step_state"`
		}

		newModelState, err := retrieveModelState(apiKey, apiEndpoint, idToken, txnId, testBase)
		if !assert.Nil(T, err) {
			return
		}

		log.Println("new model state", newModelState)

		var states []ModelState
		err = json.Unmarshal([]byte(newModelState), &states)
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
		assert.Equal(T, "completed", stateMap["ess3"])
	})

}
