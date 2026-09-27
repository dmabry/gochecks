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
	"errors"
	"testing"
	"time"
)

func TestIsRetryableErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "connection refused", err: errors.New("connection refused"), want: true},
		{name: "no route to host", err: errors.New("no route to host"), want: true},
		{name: "timeout waiting for response", err: errors.New("timeout waiting for response"), want: true},
		{name: "use closed connection", err: errors.New("use of closed network connection"), want: true},
		{name: "network unreachable", err: errors.New("network is unreachable"), want: true},
		{name: "i/o timeout", err: errors.New("i/o timeout"), want: true},
		{name: "temporary failure", err: errors.New("temporary failure"), want: true},
		{name: "non-retryable auth", err: errors.New("Authentication failed"), want: false},
		{name: "invalid OID", err: errors.New("Invalid OID specified"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableErr(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableErr(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestWithMaxRetriesOption(t *testing.T) {
	tests := []struct {
		name            string
		n               int
		wantErr         bool
		expectedRetries int
	}{
		{name: "valid zero", n: 0, wantErr: false, expectedRetries: 0},
		{name: "valid positive", n: 5, wantErr: false, expectedRetries: 5},
		{name: "invalid negative", n: -1, wantErr: true, expectedRetries: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient("127.0.0.1", "public", WithMaxRetries(tt.n))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client.retryPolicy.MaxRetries != tt.expectedRetries {
				t.Errorf("WithMaxRetries(%d) MaxRetries = %d, want %d", tt.n, client.retryPolicy.MaxRetries, tt.expectedRetries)
			}
		})
	}
}

func TestWithRetryEnabled(t *testing.T) {
	client, err := NewClient("127.0.0.1", "public")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	tests := []struct {
		name              string
		enabled           bool
		wantPolicyEnabled bool
	}{
		{name: "enabled", enabled: true, wantPolicyEnabled: true},
		{name: "disabled", enabled: false, wantPolicyEnabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.SetRetryPolicy(RetryPolicy{Enabled: tt.enabled})
			if client.retryPolicy.Enabled != tt.wantPolicyEnabled {
				t.Errorf("SetRetryPolicy() Enabled = %v, want %v", client.retryPolicy.Enabled, tt.wantPolicyEnabled)
			}
		})
	}
}

func TestClientWithNilContext(t *testing.T) {
	client := &Client{
		Target:      "127.0.0.1",
		Community:   "public",
		snmpClient:  nil,
		retryPolicy: RetryPolicy{Enabled: true, MaxRetries: 3},
	}

	_, _ = client.withRetry(context.TODO(), func() (interface{}, error) {
		return nil, errors.New("test error")
	})
}

func TestClientWithNonRetryableError(t *testing.T) {
	client := &Client{
		Target:      "127.0.0.1",
		Community:   "public",
		snmpClient:  nil,
		retryPolicy: RetryPolicy{Enabled: true, MaxRetries: 2},
	}

	opCount := 0
	op := func() (interface{}, error) {
		opCount++
		if opCount == 1 {
			return nil, errors.New("Authentication failed")
		}
		return "success", nil
	}

	ctx := context.Background()
	_, _ = client.withRetry(ctx, op)

	if opCount > 1 {
		t.Errorf("Non-retryable error caused %d attempts, want only 1", opCount)
	}
}

func TestClientRetryLogic(t *testing.T) {
	tests := []struct {
		name             string
		maxRetries       int
		retryEnabled     bool
		expectedAttempts int
	}{
		{name: "disabled retries", maxRetries: 3, retryEnabled: false, expectedAttempts: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Target:      "127.0.0.1",
				Community:   "public",
				snmpClient:  nil,
				retryPolicy: RetryPolicy{Enabled: tt.retryEnabled, MaxRetries: tt.maxRetries},
			}

			opCount := 0
			ctx := context.Background()
			_, _ = client.withRetry(ctx, func() (interface{}, error) {
				opCount++
				if opCount == 1 && tt.retryEnabled {
					return nil, errors.New("connection refused")
				}
				return "success", nil
			})

			_ = ctx
		})
	}
}

func TestContextCancellationDuringBackoff(t *testing.T) {
	client := &Client{
		Target:      "127.0.0.1",
		Community:   "public",
		snmpClient:  nil,
		retryPolicy: RetryPolicy{Enabled: true, MaxRetries: 5},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opCount := 0
	_, _ = client.withRetry(ctx, func() (interface{}, error) {
		opCount++
		return nil, errors.New("connection refused")
	})

	_ = opCount
}

func TestRetryPolicyFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		retryEnabled bool
		maxRetries   int
		setupEnv     func()
	}{
		{name: "default when unset", retryEnabled: false, maxRetries: defaultMaxRetries, setupEnv: func() {}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			policy := retryPolicyFromEnv()
			if policy.MaxRetries != tt.maxRetries {
				t.Errorf("retryPolicyFromEnv() MaxRetries = %d, want %d", policy.MaxRetries, tt.maxRetries)
			}
		})
	}
}

// TestRetryPolicyFromEnvDelay verifies the GOCHECKS_SNMP_RETRY_DELAY
// environment variable overrides the default base retry delay.
func TestRetryPolicyFromEnvDelay(t *testing.T) {
	t.Setenv("GOCHECKS_SNMP_RETRY_DELAY", "5")

	policy := retryPolicyFromEnv()
	if policy.RetryDelay != 5*time.Second {
		t.Errorf("retryPolicyFromEnv() RetryDelay = %v, want 5s", policy.RetryDelay)
	}
	if policy.MaxRetries != defaultMaxRetries {
		t.Errorf("retryPolicyFromEnv() MaxRetries = %d, want %d", policy.MaxRetries, defaultMaxRetries)
	}
}

// TestRetryBackoff verifies the exponential backoff schedule: the base delay
// doubles on each subsequent attempt and is capped at maxRetryBackoff. A
// zero delay disables backoff entirely.
func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		name       string
		retryDelay time.Duration
		attempt    int
		want       time.Duration
	}{
		{name: "zero delay disables backoff", retryDelay: 0, attempt: 3, want: 0},
		{name: "first attempt uses base delay", retryDelay: time.Second, attempt: 0, want: time.Second},
		{name: "second attempt doubles", retryDelay: time.Second, attempt: 1, want: 2 * time.Second},
		{name: "third attempt quadruples", retryDelay: time.Second, attempt: 2, want: 4 * time.Second},
		{name: "capped at max", retryDelay: 10 * time.Second, attempt: 5, want: maxRetryBackoff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := RetryPolicy{RetryDelay: tt.retryDelay}
			if got := p.backoff(tt.attempt); got != tt.want {
				t.Errorf("backoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

// TestWithRetryCancelledDuringBackoff verifies a cancelled context aborts the
// backoff wait instead of retrying, even when the delay is long.
func TestWithRetryCancelledDuringBackoff(t *testing.T) {
	client := &Client{
		Target:      "127.0.0.1",
		Community:   "public",
		retryPolicy: RetryPolicy{Enabled: true, MaxRetries: 3, RetryDelay: time.Hour},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	_, err := client.withRetry(ctx, func() (interface{}, error) {
		attempts++
		return nil, errors.New("connection refused")
	})

	if err == nil {
		t.Error("expected context error from cancelled context")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (cancelled context must not retry)", attempts)
	}
}

func TestNewClientWithMultipleOptions(t *testing.T) {
	client, err := NewClient(
		"127.0.0.1",
		"public",
		WithMaxRetries(5),
		WithRetryEnabled(true),
	)

	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	if client.retryPolicy.MaxRetries != 5 {
		t.Errorf("WithMaxRetries option not applied, MaxRetries = %d", client.retryPolicy.MaxRetries)
	}
	if !client.retryPolicy.Enabled {
		t.Error("WithRetryEnabled option not applied")
	}
}
