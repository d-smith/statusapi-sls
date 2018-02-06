package api

import (
	"bytes"
	"fmt"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

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
		//T.Skip() // pending
	})

	When(`^I retrieve the model state for the correlated events$`, func() {
		//T.Skip() // pending
	})

	Then(`^the state of the model reflects the events$`, func() {
		//T.Skip() // pending
	})

}
