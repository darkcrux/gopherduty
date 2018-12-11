package gopherduty

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type sampleDetail struct {
	Data1 string
	Data2 []string
}

func TestBackOffDelay(t *testing.T) {
	pd := &PagerDuty{
		MaxRetry:          3,
		RetryBaseInterval: 1,
	}

	delays := []int{1, 2, 4}

	for i := 0; i < pd.MaxRetry; i++ {
		now := time.Now()
		pd.retries = i
		pd.delayRetry()
		actualDelay := int(time.Since(now).Seconds())
		expectedDelay := delays[i]
		if actualDelay != expectedDelay {
			t.Errorf("expected delay is %d, actual delay is %d", expectedDelay, actualDelay)
		}
	}
}

func TestRetryOnRequest(t *testing.T) {
	pd := &PagerDuty{
		MaxRetry:          3,
		RetryBaseInterval: 1,
	}

	expectedRuntime := 7
	now := time.Now()
	response := pd.Trigger("", "", "", "", nil)
	actual := int(time.Since(now).Seconds())
	if !response.HasErrors() {
		t.Error("This should have been an error")
	}
	if actual < expectedRuntime {
		t.Errorf("Expected runtime is %d, actual is %d", expectedRuntime, actual)
	}

}

func TestRetryOnRatelimit(t *testing.T) {
	rateLimitErrorMsg := `{
		"status": "throttle exceeded",
		"message": "Requests for this service are arriving too quickly.  Please retry later."
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, rateLimitErrorMsg, http.StatusForbidden)
	}))
	defer ts.Close()

	pd := &PagerDuty{
		MaxRetry:          3,
		RetryBaseInterval: 1,
		endpoint:          &ts.URL,
	}

	expectedRuntime := 7
	now := time.Now()
	response := pd.Trigger("", "", "", "", nil)
	actual := int(time.Since(now).Seconds())
	fmt.Printf("Example error message: %v\n", response.Errors)
	if !response.HasErrors() {
		t.Error("This should have been an error")
	}
	if actual < expectedRuntime {
		t.Errorf("Expected runtime is %d, actual is %d", expectedRuntime, actual)
	}
}
