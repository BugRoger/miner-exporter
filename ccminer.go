package main

import (
	"bufio"
	"net"
	"strconv"
	"strings"
)

type CCMinerAPI interface {
	Summary() (string, error)
	Threads() (string, error)
	Pool() (string, error)
}

type CCMinerClient struct {
	api CCMinerAPI
}

type client struct {
	address string
}

func NewCCMinerClient(address string) *CCMinerClient {
	return &CCMinerClient{&client{address}}
}

func (c *CCMinerClient) Name() string {
	return "ccminer"
}

func (c *CCMinerClient) Collect() (*Metrics, error) {
	resp, err := c.api.Summary()
	if err != nil {
		return nil, err
	}
	summary := toMap(resp)

	resp, err = c.api.Pool()
	if err != nil {
		return nil, err
	}
	pool := toMap(resp)

	resp, err = c.api.Threads()
	if err != nil {
		return nil, err
	}
	threads := toMaps(resp)

	uptime, _ := strconv.ParseFloat(summary["UPTIME"], 64)
	accepted, _ := strconv.ParseFloat(pool["ACC"], 64)
	rejected, _ := strconv.ParseFloat(pool["REJ"], 64)
	stale, _ := strconv.ParseFloat(pool["STALE"], 64)

	byGPU := []float64{}
	total := 0.0
	for _, gpu := range threads {
		rate, _ := strconv.ParseFloat(gpu["KHS"], 64)
		byGPU = append(byGPU, rate)
		total = total + rate
	}

	return &Metrics{
		Version: summary["VER"],
		Uptime:  uptime,
		Algorithms: []Algorithm{
			{
				Name: summary["ALGO"],
				Shares: Shares{
					Accepted: accepted,
					Rejected: rejected,
					Stale:    stale,
				},
				Rates: Rates{
					Total: total,
					ByGPU: byGPU,
				},
			},
		},
	}, nil
}

func (c *client) rpc(command string) (string, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return "", err
	}

	_, err = conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(bufio.NewReader(conn))
	if scanner.Scan(); scanner.Err() != nil {
		return "", err
	}

	return scanner.Text(), nil
}

func (c *client) Summary() (string, error) {
	return c.rpc("summary")
}

func (c *client) Threads() (string, error) {
	return c.rpc("threads")
}

func (c *client) Pool() (string, error) {
	return c.rpc("pool")
}

func toMaps(input string) []map[string]string {
	result := []map[string]string{}

	for _, value := range strings.Split(input, "|") {
		inner := toMap(value)
		if len(inner) > 0 {
			result = append(result, inner)
		}
	}

	return result
}

func toMap(input string) map[string]string {
	result := map[string]string{}

	for _, value := range strings.Split(input, ";") {
		split := strings.Split(value, "=")
		if len(split) == 2 {
			result[split[0]] = split[1]
		}
	}

	return result
}
