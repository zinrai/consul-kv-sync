package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	MaxOpsPerTransaction = 64
)

// ConsulClient represents a client for Consul API
type ConsulClient struct {
	addr       string
	datacenter string
	httpClient *http.Client
}

// NewConsulClient creates a new Consul client
func NewConsulClient(addr, datacenter string) *ConsulClient {
	return &ConsulClient{
		addr:       addr,
		datacenter: datacenter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// executeTransaction executes a transaction with the given operations
func (c *ConsulClient) executeTransaction(ops []TxnOp) (*TxnResponse, error) {
	url := fmt.Sprintf("%s/v1/txn?dc=%s", c.addr, c.datacenter)

	jsonData, err := json.Marshal(ops)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operations: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response regardless of status code
	var txnResp TxnResponse
	if err := json.Unmarshal(body, &txnResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		return &txnResp, nil
	case http.StatusConflict:
		return &txnResp, fmt.Errorf("transaction rolled back with %d errors", len(txnResp.Errors))
	default:
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
}

// syncKVPairs synchronizes all KV pairs to Consul
func (c *ConsulClient) syncKVPairs(pairs []KVPair, verbose bool) (*ExecutionSummary, error) {
	ops := createTransactionOps(pairs)
	chunks := chunkOps(ops, MaxOpsPerTransaction)

	summary := &ExecutionSummary{
		TotalKeys:    len(pairs),
		TotalBatches: len(chunks),
		Results:      make([]BatchResult, 0, len(chunks)),
	}

	if verbose {
		fmt.Printf("Syncing %d key-value pairs in %d batches...\n", len(pairs), len(chunks))
	}

	for i, chunk := range chunks {
		if verbose {
			fmt.Printf("Processing batch %d/%d (%d operations)...\n", i+1, len(chunks), len(chunk))
		}

		result := BatchResult{
			BatchIndex:   i,
			ProcessedOps: len(chunk),
		}

		txnResp, err := c.executeTransaction(chunk)
		if err != nil {
			result.Success = false
			result.Error = err
			if txnResp != nil {
				result.OpErrors = txnResp.Errors
			}
			summary.FailedBatches++
		} else {
			result.Success = true
			summary.SuccessBatches++
		}

		summary.Results = append(summary.Results, result)

		// Add a small delay between batches to avoid overwhelming the server
		if i < len(chunks)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return summary, nil
}

// formatExecutionSummary formats the execution summary for display
func formatExecutionSummary(summary *ExecutionSummary) string {
	var sb strings.Builder

	writeHeader(&sb)
	writeSummaryStats(&sb, summary)

	if summary.FailedBatches > 0 {
		writeFailedBatches(&sb, summary)
	}

	writeStatusMessage(&sb, summary)

	return sb.String()
}

func writeHeader(sb *strings.Builder) {
	sb.WriteString("\n" + strings.Repeat("=", 60) + "\n")
	sb.WriteString("Execution Summary\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")
}

func writeSummaryStats(sb *strings.Builder, summary *ExecutionSummary) {
	sb.WriteString(fmt.Sprintf("Total key-value pairs: %d\n", summary.TotalKeys))
	sb.WriteString(fmt.Sprintf("Total batches: %d\n", summary.TotalBatches))
	sb.WriteString(fmt.Sprintf("Successful batches: %d\n", summary.SuccessBatches))
	sb.WriteString(fmt.Sprintf("Failed batches: %d\n", summary.FailedBatches))
}

func writeFailedBatches(sb *strings.Builder, summary *ExecutionSummary) {
	sb.WriteString("\nFailed Batches:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, result := range summary.Results {
		if !result.Success {
			sb.WriteString(fmt.Sprintf("Batch %d: %v\n", result.BatchIndex+1, result.Error))
			writeOperationErrors(sb, result.OpErrors)
		}
	}
}

func writeOperationErrors(sb *strings.Builder, errors []TxnError) {
	for _, opErr := range errors {
		sb.WriteString(fmt.Sprintf("  - Operation %d: %s\n", opErr.OpIndex, opErr.What))
	}
}

func writeStatusMessage(sb *strings.Builder, summary *ExecutionSummary) {
	switch {
	case summary.SuccessBatches == summary.TotalBatches:
		sb.WriteString("\n[SUCCESS] All operations completed successfully!\n")
	case summary.SuccessBatches > 0:
		sb.WriteString("\n[WARNING] Partial success: Some batches failed.\n")
	default:
		sb.WriteString("\n[ERROR] All batches failed.\n")
	}
}
