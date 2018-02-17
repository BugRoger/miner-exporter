package main

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
)

type ClaymoreDualMinerClient struct {
	network string
	address string
}

func NewClaymoreDualMinerClient(network string, address string) *ClaymoreDualMinerClient {
	return &ClaymoreDualMinerClient{network, address}
}

func (c *ClaymoreDualMinerClient) Name() string {
	return "ClaymoreDualMiner"
}

func (m *ClaymoreDualMinerClient) Collect() (*Metrics, error) {
	client, err := m.client()
	if err != nil {
		return nil, err
	}

	reply := []string{}
	if err = client.Call("miner_getstat1", nil, &reply); err != nil {
		return nil, err
	}

	return m.parse(reply), nil
}

func (m *ClaymoreDualMinerClient) client() (*rpc.Client, error) {
	conn, err := net.Dial(m.network, m.address)
	if err != nil {
		return nil, err
	}

	return jsonrpc.NewClient(conn), nil
}

func (m *ClaymoreDualMinerClient) parse(reply []string) *Metrics {
	version := reply[0]
	uptime := parseGarble(reply[1])[0] * 60
	eth := parseGarble(reply[2])
	ethRates := parseGarble(reply[3])
	alt := parseGarble(reply[4])
	altRates := parseGarble(reply[5])

	return &Metrics{
		Version: version,
		Uptime:  uptime,
		Algorithms: []Algorithm{
			Algorithm{
				Name: "daggerhashimoto",
				Shares: Shares{
					Accepted: eth[1],
					Rejected: eth[2],
				},
				Rates: Rates{
					Total: eth[0],
					ByGPU: ethRates,
				},
			},
			Algorithm{
				Name: "decred",
				Shares: Shares{
					Accepted: alt[1],
					Rejected: alt[2],
				},
				Rates: Rates{
					Total: alt[0],
					ByGPU: altRates,
				},
			},
		},
	}
}

func unzip(i []string) ([]string, []string) {
	a := []string{}
	b := []string{}

	for len(i) > 1 {
		a = append(a, i[0])
		b = append(b, i[1])
		i = i[2:]
	}

	return a, b
}

func parseZippedGarble(input string) ([]float64, []float64) {
	a, b := unzip(strings.Split(input, ";"))
	return toFloat(a...), toFloat(b...)
}

func parseGarble(input string) []float64 {
	return toFloat(strings.Split(input, ";")...)
}

func toFloat(input ...string) []float64 {
	r := []float64{}
	for _, i := range input {
		f, _ := strconv.ParseFloat(i, 64)
		r = append(r, f)
	}
	return r
}
