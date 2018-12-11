package gopherduty

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

const defaultEndpoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

type pagerDutyRequest struct {
	ServiceKey  string      `json:"service_key"`
	EventType   string      `json:"event_type"`
	IncidentKey string      `json:"incident_key,omitempty"`
	Description string      `json:"description"`
	Client      string      `json:"client,omitempty"`
	ClientUrl   string      `json:"client_url,omitempty"`
	Details     interface{} `json:"details"`
}

func (p *pagerDutyRequest) submit(alternateEndpoint *string) (pagerResponse *PagerDutyResponse) {
	pagerResponse = &PagerDutyResponse{}

	body, err := json.Marshal(p)
	if err != nil {
		pagerResponse.appendError(err)
		return pagerResponse
	}

	endpoint := defaultEndpoint
	if alternateEndpoint != nil {
		endpoint = *alternateEndpoint
	}

	buf := bytes.NewBuffer(body)
	response, err := http.Post(endpoint, "application/json", buf)
	if err != nil {
		pagerResponse.appendError(err)
		return pagerResponse
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		pagerResponse.appendError(err)
		return pagerResponse
	}

	pagerResponse.parse(responseBody)

	// PagerDuty sends a HTTP 403 when you have been rate-limited
	// rate-limiting implies this current request has not been received.
	if response.StatusCode == http.StatusForbidden {
		errMsg := fmt.Sprintf("%v. %v", pagerResponse.Status, pagerResponse.Message)
		pagerResponse.appendError(errors.New(errMsg))
		return pagerResponse
	}

	return pagerResponse
}
