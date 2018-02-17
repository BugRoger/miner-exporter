package main

import (
	"bufio"
	"encoding/json"
	"net"
)

type DSTMClient struct {
	api DSTMAPI
}

type DSTMAPI interface {
	GetStat() (*getStat, error)
}

type getStat struct {
	Id      int      `json:"id"`
	Uptime  int      `json:"uptime"`
	Contime int      `json:"contime"`
	Server  string   `json:"server"`
	Port    int      `json:"port"`
	User    string   `json:"user"`
	Version string   `json:"version"`
	Error   *string  `json:"error"`
	Result  []result `json:"result"`
}

type result struct {
	GpuID          int     `json:"gpu_id"`
	Temperature    int     `json:"temperature"`
	SolPerSecond   float64 `json:"sol_ps"`
	AvgSolPerSec   float64 `json:"avg_sol_ps"`
	SolPerWatt     float64 `json:"sol_pw"`
	AvgSolPerWatt  float64 `json:"avg_sol_pw"`
	PowerUsage     float64 `json:"power_usage"`
	AvgPowerUsage  float64 `json:"avg_power_uasge"`
	AcceptedShares int     `json:"accepted_shares"`
	RejectedShares int     `json:"rejected_shares"`
	Latency        int     `json:"latency"`
}

type dstmAPIClient struct {
	address string
}

func (c dstmAPIClient) GetStat() (*getStat, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, err
	}
	result := getStat{}
	command := "{\"id\": 1, \"method\": \"getstat\"}"
	_, err = conn.Write([]byte(command))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bufio.NewReader(conn))
	if scanner.Scan(); scanner.Err() != nil {
		return nil, err
	}

	json.Unmarshal([]byte(scanner.Text()), &result)

	return &result, nil
}

func NewDSTMClient(address string) *DSTMClient {
	return &DSTMClient{dstmAPIClient{address}}
}

func (c *DSTMClient) Name() string {
	return "dstm"
}

func (c *DSTMClient) Collect() (*Metrics, error) {
	stats, err := c.api.GetStat()
	if err != nil {
		return nil, err
	}

	byGPU := []float64{}
	accepted := 0
	rejected := 0
	total := 0.0
	for _, gpu := range stats.Result {
		rate := gpu.SolPerSecond
		byGPU = append(byGPU, rate)
		total = total + rate
		accepted = accepted + gpu.AcceptedShares
		rejected = rejected + gpu.RejectedShares
	}

	return &Metrics{
		Version: stats.Version,
		Uptime:  float64(stats.Uptime),
		Algorithms: []Algorithm{
			{
				Name: "equihash",
				Shares: Shares{
					Accepted: float64(accepted),
					Rejected: float64(rejected),
					Stale:    0.0,
				},
				Rates: Rates{
					Total: total,
					ByGPU: byGPU,
				},
			},
		},
	}, nil
}
