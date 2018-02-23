package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d-smith/statusapi-sls/event"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func retrieveTxnEvents(apiKey, apiEndpoint, idToken, txnId string) ([]event.StatusEvent, error) {
	//curl -H "x-api-key: XXXX"  'https://ENDPOINT/dev/status/api/v1/instances/1a?model=model1'
	//https://oou3pdrtw2.execute-api.us-east-1.amazonaws.com/dev/status/api/v1/instances/{id}
	requestUrl := fmt.Sprintf("https://%s/dev/status/api/v1/instances/%s", apiEndpoint, txnId)
	log.Println("retrieveTxnEvents: get", requestUrl)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
//	req.Header.Add("x-api-key", apiKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

	log.Println("auth header", req.Header.Get("Authorization"))

	client := &http.Client{}
	log.Println("make test request")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error on test request: %s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http call status not ok: %d", resp.StatusCode))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println("Body --> ", string(body))
	var statusEvents []event.StatusEvent

	err = json.Unmarshal([]byte(body), &statusEvents)

	return statusEvents, err
}

func init() {

	var (
		apiKey      = os.Getenv("APIKEY")
		apiEndpoint = os.Getenv("API_ENDPOINT")
		idToken = os.Getenv("STATUS_ID_TOKEN")
		txnId       string
	)

	if apiKey == "" || apiEndpoint == "" {
		log.Println("Must set both APIKEY and API_ENDPOINT environment variables to run gucumber tests")
		os.Exit(1)
	}

	When(`^I post events for a transaction$`, func() {
		var err error
		txnId, err = postEventsForModel(apiKey, apiEndpoint, idToken)
		if !assert.Nil(T, err) {
			log.Printf(err.Error())
			return
		}
	})

	Then(`^I can retrieve those events using the transaction id$`, func() {
		statusEvents, err := retrieveTxnEvents(apiKey, apiEndpoint, idToken, txnId)
		if assert.Nil(T, err) {
			log.Printf("%v", statusEvents)
			assert.Equal(T, 2, len(statusEvents))

			for i := 0; i < 2; i++ {
				se := statusEvents[i]
				switch se.EventId {
				case "0":
					assert.Equal(T, txnId, se.TransactionId)
					assert.Equal(T, "0", se.EventId)
					assert.Equal(T, "s1", se.Step)
					assert.Equal(T, "completed", se.StepState)
				case "1":
					assert.Equal(T, txnId, se.TransactionId)
					assert.Equal(T, "1", se.EventId)
					assert.Equal(T, "s2", se.Step)
					assert.Equal(T, "completed", se.StepState)
				default:
					assert.Fail(T, "Status events include unexpected event id")
				}
			}
		}
	})

}
