package main

// Config represents the environment configuration
type Config struct {
	Environments map[string][]string `yaml:",inline"`
}

// KVPair represents a key-value pair
type KVPair struct {
	Key   string
	Value string
}

// TxnKVOp represents a KV operation in Consul transaction
type TxnKVOp struct {
	Verb  string `json:"Verb"`
	Key   string `json:"Key"`
	Value string `json:"Value"`
	Flags uint64 `json:"Flags,omitempty"`
}

// TxnOp represents a transaction operation
type TxnOp struct {
	KV *TxnKVOp `json:"KV,omitempty"`
}

// TxnResponse represents the response from Consul transaction API
type TxnResponse struct {
	Results []TxnResult `json:"Results"`
	Errors  []TxnError  `json:"Errors"`
}

// TxnResult represents a successful operation result
type TxnResult struct {
	KV *KVData `json:"KV,omitempty"`
}

// KVData represents KV data in the response
type KVData struct {
	LockIndex   uint64 `json:"LockIndex"`
	Key         string `json:"Key"`
	Flags       uint64 `json:"Flags"`
	Value       string `json:"Value"`
	CreateIndex uint64 `json:"CreateIndex"`
	ModifyIndex uint64 `json:"ModifyIndex"`
}

// TxnError represents an error in transaction
type TxnError struct {
	OpIndex int    `json:"OpIndex"`
	What    string `json:"What"`
}

// DuplicateInfo contains information about duplicate keys
type DuplicateInfo struct {
	Key   string
	Files []FileSource
}

// FileSource represents the source file and value of a key
type FileSource struct {
	Filename string
	Value    interface{}
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	BatchIndex   int
	Success      bool
	Error        error
	OpErrors     []TxnError
	ProcessedOps int
}

// ExecutionSummary represents the overall execution summary
type ExecutionSummary struct {
	TotalKeys      int
	TotalBatches   int
	SuccessBatches int
	FailedBatches  int
	Results        []BatchResult
}
