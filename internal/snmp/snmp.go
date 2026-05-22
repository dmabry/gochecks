/*
   Copyright 2024 David Mabry

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package snmp

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"log"
)

// timeout15 is a constant representing a timeout duration of 15 seconds.
const (
	timeout15         = time.Duration(15) * time.Second
	defaultMaxRetries = 3
)

// RetryPolicy defines the retry behavior for SNMP operations.
type RetryPolicy struct {
	Enabled    bool
	MaxRetries int
}

// Client represents an SNMP client that allows connecting to a target SNMP device.
type Client struct {
	Target      string
	Community   string
	snmpClient  *gosnmp.GoSNMP
	retryPolicy RetryPolicy
}

// ClientOption is a functional option for configuring an SNMP Client.
type ClientOption func(*Client) error

// NewClient creates a new SNMP client with the given target and community string,
// applying any provided configuration options.
func NewClient(target, community string, opts ...ClientOption) (*Client, error) {
	client := &Client{
		Target:      target,
		Community:   community,
		retryPolicy: retryPolicyFromEnv(),
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// WithMaxRetries sets the maximum number of retries for SNMP operations.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) error {
		if n < 0 {
			return &OptionError{MaxRetries: n}
		}
		c.retryPolicy.MaxRetries = n
		return nil
	}
}

// WithRetryEnabled enables or disables retry behavior for SNMP operations.
func WithRetryEnabled(enabled bool) ClientOption {
	return func(c *Client) error {
		c.retryPolicy.Enabled = enabled
		return nil
	}
}

// OptionError is returned when a ClientOption receives invalid input.
type OptionError struct {
	MaxRetries int
}

func (e *OptionError) Error() string {
	return "invalid option: MaxRetries must be >= 0, got " + strconv.Itoa(e.MaxRetries)
}

// SetRetryPolicy updates the retry policy for the client.
func (c *Client) SetRetryPolicy(policy RetryPolicy) {
	c.retryPolicy = policy
}

// retryPolicyFromEnv reads retry configuration from environment variables.
// GOCHECKS_SNMP_RETRY_ENABLED controls whether retries are enabled (default: false).
// GOCHECKS_SNMP_MAX_RETRIES sets the maximum retry count (default: 3).
func retryPolicyFromEnv() RetryPolicy {
	policy := RetryPolicy{
		Enabled:    false,
		MaxRetries: defaultMaxRetries,
	}

	if v := os.Getenv("GOCHECKS_SNMP_RETRY_ENABLED"); v != "" {
		policy.Enabled = strings.ToLower(v) == "true" || v == "1"
	}

	if v := os.Getenv("GOCHECKS_SNMP_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			policy.MaxRetries = n
		}
	}

	return policy
}

// isRetryableErr returns true if the given error represents a transient network
// condition that may resolve on retry.
func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"connection refused",
		"no route to host",
		"timeout waiting for response",
		"use of closed network connection",
		"network is unreachable",
		"i/o timeout",
		"temporary failure",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	return false
}

// withRetry executes the given operation with optional retry logic.
// If the context is nil, it is treated as context.Background().
// If retries are disabled or the error is non-retryable, it returns immediately.
func (c *Client) withRetry(ctx context.Context, op func() (interface{}, error)) (interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !c.retryPolicy.Enabled {
		return op()
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		result, err := op()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !isRetryableErr(err) {
			return nil, err
		}

		if attempt >= c.retryPolicy.MaxRetries {
			break
		}

		// Check if context is cancelled before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return nil, lastErr
}

// createGoSNMP creates and connects a new gosnmp.GoSNMP instance.
func (s *Client) createGoSNMP() (*gosnmp.GoSNMP, error) {
	snmpClient := &gosnmp.GoSNMP{
		Target:    s.Target,
		Port:      161,
		Community: s.Community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout15,
	}

	if err := snmpClient.Connect(); err != nil {
		return nil, err
	}

	return snmpClient, nil
}

// Connect establishes a connection to the SNMP target with context support.
// If the context is nil, it is treated as context.Background().
// The connection attempt is retried according to the client's retry policy.
//
// Example usage:
//
//	err := client.Connect(ctx)
//	if err != nil {
//
// Connect establishes a connection to the SNMP target with context support.
// If the context is nil, it is treated as context.Background().
// The connection attempt is retried according to the client's retry policy.
func (s *Client) Connect(ctx context.Context) error {
	_, err := s.withRetry(ctx, func() (interface{}, error) {
		snmpClient, err := s.createGoSNMP()
		if err != nil {
			return nil, err
		}
		s.snmpClient = snmpClient
		return snmpClient, nil
	})
	return err
}

// Close closes the underlying SNMP connection.
func (s *Client) Close() error {
	if s.snmpClient != nil && s.snmpClient.Conn != nil {
		return s.snmpClient.Conn.Close()
	}
	return nil
}

// GetValue retrieves SNMP values for the given OIDs using the client's connection.
// It returns the SNMP packet containing the result values, the duration of the SNMP request,
// and any error encountered during the process.
// If the context is nil, it is treated as context.Background().
func (s *Client) GetValue(ctx context.Context, oids []string) (*gosnmp.SnmpPacket, time.Duration, error) {
	var result *gosnmp.SnmpPacket
	var latency time.Duration

	_, err := s.withRetry(ctx, func() (interface{}, error) {
		snmpClient, err := s.createGoSNMP()
		if err != nil {
			return nil, err
		}
		defer func() {
			if closeErr := snmpClient.Conn.Close(); closeErr != nil {
				log.Printf("Error closing SNMP connection: %v", closeErr)
			}
		}()

		start := time.Now()
		pkt, err := snmpClient.Get(oids)
		if err != nil {
			return nil, err
		}

		latency = time.Since(start)
		result = pkt
		return pkt, nil
	})

	return result, latency, err
}

// Walk retrieves SNMP tree for the given OID using the client's connection.
// It returns a map with the OID as the key and its value as the value,
// the duration of the SNMP request, and any error encountered during the process.
// If the context is nil, it is treated as context.Background().
func (s *Client) Walk(ctx context.Context, baseOid string) (map[string]interface{}, time.Duration, error) {
	var result map[string]interface{}
	var latency time.Duration

	_, err := s.withRetry(ctx, func() (interface{}, error) {
		snmpClient, err := s.createGoSNMP()
		if err != nil {
			return nil, err
		}
		defer func() {
			if closeErr := snmpClient.Conn.Close(); closeErr != nil {
				log.Printf("Error closing SNMP connection: %v", closeErr)
			}
		}()

		start := time.Now()
		oidValues := make(map[string]interface{})

		err = snmpClient.BulkWalk(baseOid, func(pdu gosnmp.SnmpPDU) error {
			oidValues[pdu.Name] = pdu.Value
			return nil
		})
		if err != nil {
			return nil, err
		}

		latency = time.Since(start)
		result = oidValues
		return oidValues, nil
	})

	return result, latency, err
}
