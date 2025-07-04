package post2post

import (
	"fmt"
	"strings"
	"time"
)

// HelloWorldProcessor always returns "Hello World" message
type HelloWorldProcessor struct{}

func (h *HelloWorldProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	return map[string]interface{}{
		"message":    "Hello World",
		"request_id": requestID,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05 MST"),
	}, nil
}

// EchoProcessor simply returns the original payload with additional metadata
type EchoProcessor struct{}

func (e *EchoProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	return map[string]interface{}{
		"original_payload": payload,
		"request_id":       requestID,
		"processed_at":     time.Now().Format("2006-01-02 15:04:05 MST"),
		"processor":        "echo",
		"status":           "echoed",
	}, nil
}

// TimestampProcessor adds timestamp information to the payload
type TimestampProcessor struct{}

func (t *TimestampProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	now := time.Now()
	return map[string]interface{}{
		"data":          payload,
		"request_id":    requestID,
		"processed_at":  now.Format("2006-01-02 15:04:05 MST"),
		"unix_time":     now.Unix(),
		"processor":     "timestamp",
		"day_of_week":   now.Weekday().String(),
		"processing_ms": 100, // Simulated processing time
	}, nil
}

// CounterProcessor maintains a counter and includes it in responses
type CounterProcessor struct {
	count int
}

func NewCounterProcessor() *CounterProcessor {
	return &CounterProcessor{count: 0}
}

func (c *CounterProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	c.count++
	return map[string]interface{}{
		"payload":       payload,
		"request_id":    requestID,
		"count":         c.count,
		"processed_at":  time.Now().Format("2006-01-02 15:04:05 MST"),
		"processor":     "counter",
		"message":       fmt.Sprintf("This is request number %d", c.count),
	}, nil
}

// AdvancedContextProcessor demonstrates using the advanced context interface
type AdvancedContextProcessor struct {
	ServiceName string
}

func NewAdvancedContextProcessor(serviceName string) *AdvancedContextProcessor {
	return &AdvancedContextProcessor{ServiceName: serviceName}
}

// Process implements PayloadProcessor interface as a fallback
func (a *AdvancedContextProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	// Create minimal context for basic interface
	context := ProcessorContext{
		RequestID:  requestID,
		ReceivedAt: time.Now(),
	}
	return a.ProcessWithContext(payload, context)
}

func (a *AdvancedContextProcessor) ProcessWithContext(payload interface{}, context ProcessorContext) (interface{}, error) {
	processingTime := time.Since(context.ReceivedAt)
	
	response := map[string]interface{}{
		"service_name":     a.ServiceName,
		"original_payload": payload,
		"context": map[string]interface{}{
			"request_id":     context.RequestID,
			"callback_url":   context.URL,
			"received_at":    context.ReceivedAt.Format("2006-01-02 15:04:05.000 MST"),
			"processing_ms":  processingTime.Nanoseconds() / 1000000,
		},
		"processed_at": time.Now().Format("2006-01-02 15:04:05 MST"),
		"processor":    "advanced_context",
		"status":       "processed_with_context",
	}
	
	// Add Tailscale info if present
	if context.TailnetKey != "" {
		response["tailscale"] = map[string]interface{}{
			"enabled":     true,
			"key_prefix":  context.TailnetKey[:min(len(context.TailnetKey), 10)] + "...",
			"secure_mode": true,
		}
	}
	
	return response, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TransformProcessor transforms string payloads to uppercase
type TransformProcessor struct{}

func (t *TransformProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	response := map[string]interface{}{
		"request_id":   requestID,
		"processed_at": time.Now().Format("2006-01-02 15:04:05 MST"),
		"processor":    "transform",
	}
	
	// Transform based on payload type
	switch v := payload.(type) {
	case string:
		response["transformed"] = strings.ToUpper(v)
		response["original"] = v
		response["transformation"] = "uppercase"
		
	case map[string]interface{}:
		transformed := make(map[string]interface{})
		for key, value := range v {
			if str, ok := value.(string); ok {
				transformed[key] = strings.ToUpper(str)
			} else {
				transformed[key] = value
			}
		}
		response["transformed"] = transformed
		response["original"] = v
		response["transformation"] = "uppercase_strings"
		
	default:
		response["original"] = payload
		response["transformed"] = payload
		response["transformation"] = "no_transformation"
		response["message"] = "Only strings and maps with string values are transformed"
	}
	
	return response, nil
}

// ValidatorProcessor validates payloads and returns validation results
type ValidatorProcessor struct {
	RequiredFields []string
}

func NewValidatorProcessor(requiredFields []string) *ValidatorProcessor {
	return &ValidatorProcessor{RequiredFields: requiredFields}
}

func (v *ValidatorProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	response := map[string]interface{}{
		"request_id":   requestID,
		"processed_at": time.Now().Format("2006-01-02 15:04:05 MST"),
		"processor":    "validator",
		"original":     payload,
	}
	
	// Validate the payload
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		missing := []string{}
		present := []string{}
		
		for _, field := range v.RequiredFields {
			if _, exists := payloadMap[field]; exists {
				present = append(present, field)
			} else {
				missing = append(missing, field)
			}
		}
		
		response["validation"] = map[string]interface{}{
			"valid":           len(missing) == 0,
			"required_fields": v.RequiredFields,
			"present_fields":  present,
			"missing_fields":  missing,
		}
		
		if len(missing) == 0 {
			response["status"] = "valid"
			response["message"] = "All required fields are present"
		} else {
			response["status"] = "invalid"
			response["message"] = fmt.Sprintf("Missing required fields: %v", missing)
		}
	} else {
		response["validation"] = map[string]interface{}{
			"valid":   false,
			"message": "Payload must be a JSON object for validation",
		}
		response["status"] = "invalid"
		response["message"] = "Payload is not a JSON object"
	}
	
	return response, nil
}

// ChainProcessor allows chaining multiple processors together
type ChainProcessor struct {
	Processors []PayloadProcessor
}

func NewChainProcessor(processors ...PayloadProcessor) *ChainProcessor {
	return &ChainProcessor{Processors: processors}
}

func (c *ChainProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
	currentPayload := payload
	
	for i, processor := range c.Processors {
		result, err := processor.Process(currentPayload, requestID)
		if err != nil {
			return map[string]interface{}{
				"error":        fmt.Sprintf("Processor %d failed: %v", i, err),
				"request_id":   requestID,
				"processor":    "chain",
				"failed_at":    i,
				"processed_at": time.Now().Format("2006-01-02 15:04:05 MST"),
			}, nil
		}
		currentPayload = result
	}
	
	return map[string]interface{}{
		"result":       currentPayload,
		"request_id":   requestID,
		"processor":    "chain",
		"chain_length": len(c.Processors),
		"processed_at": time.Now().Format("2006-01-02 15:04:05 MST"),
	}, nil
}