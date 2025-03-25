package command

import (
	"sync"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplate(t *testing.T) {
	command := &CostumCommand{
		Param: map[string]interface{}{
			"username": "testUser",
		},
	}

	commandInfo := &CommandInfo{
		c: command,
	}

	tpl := "Hello {{.username}}!"
	result, err := commandInfo.getTemplate(tpl)
	assert.Nil(t, err)
	assert.Equal(t, "Hello testUser!", result)
}

func TestFetchURL_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock HTTP request
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `{"status":"ok"}`))

	// Create mock HTTPParam
	httpParam := &HTTPParam{
		URL:     "https://example.com",
		Method:  "GET",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    "",
	}

	commandInfo := &CommandInfo{
		c: &CostumCommand{
			Param: map[string]interface{}{
				"username": "testUser",
			},
		},
	}
	var wg sync.WaitGroup
	resultChan := make(chan map[string]interface{}, 1)
	wg.Add(1)

	// Execute the function
	go commandInfo.fetchURL(httpParam, &wg, resultChan, "testKey", "")

	// Wait for the result and check the response
	wg.Wait()
	close(resultChan)
	result := <-resultChan

	// Check if the result is as expected
	assert.Equal(t, map[string]interface{}{"testKey": map[string]interface{}{"status": "ok"}}, result)
}

func TestConcurrentHTTPRequests(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock HTTP request
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `{"status":"ok"}`))

	// Create mock tasks
	task := &Task{
		Name: "TestTask",
		HTTPParam: &HTTPParam{
			URL:    "https://example.com",
			Method: "GET",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "",
		},
	}

	command := &CostumCommand{
		Param: map[string]interface{}{},
		Chains: []*Chain{
			{
				Type:  TaskTypeHTTP,
				Tasks: []*Task{task},
			},
		},
	}

	commandInfo := &CommandInfo{
		c: command,
	}

	// Execute concurrent HTTP requests
	commandInfo.concurrentHTTPRequests(command.Chains[0].Tasks, "")

	// After requests are sent, check if the Param has been updated with the expected response
	assert.Equal(t, map[string]interface{}{"TestTask": map[string]interface{}{"status": "ok"}}, command.Param)
}
