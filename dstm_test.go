package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	GETSTAT = `{
   "id":1,
   "result":[
      {
         "gpu_id":0,
         "temperature":59,
         "sol_ps":422.49,
         "avg_sol_ps":430.71,
         "sol_pw":4.32,
         "avg_sol_pw":4.32,
         "power_usage":97.86,
         "avg_power_usage":99.68,
         "accepted_shares":1,
         "rejected_shares":1,
         "latency":287
      },
      {
         "gpu_id":1,
         "temperature":54,
         "sol_ps":427.96,
         "avg_sol_ps":427.46,
         "sol_pw":4.29,
         "avg_sol_pw":4.29,
         "power_usage":99.74,
         "avg_power_usage":99.55,
         "accepted_shares":0,
         "rejected_shares":0,
         "latency":0
      },
      {
         "gpu_id":2,
         "temperature":56,
         "sol_ps":431.08,
         "avg_sol_ps":428.32,
         "sol_pw":4.29,
         "avg_sol_pw":4.29,
         "power_usage":100.42,
         "avg_power_usage":99.75,
         "accepted_shares":2,
         "rejected_shares":1,
         "latency":288
      },
      {
         "gpu_id":3,
         "temperature":60,
         "sol_ps":424.16,
         "avg_sol_ps":422.06,
         "sol_pw":4.28,
         "avg_sol_pw":4.24,
         "power_usage":99.14,
         "avg_power_usage":99.51,
         "accepted_shares":5,
         "rejected_shares":0,
         "latency":293
      },
      {
         "gpu_id":4,
         "temperature":58,
         "sol_ps":421.73,
         "avg_sol_ps":427.81,
         "sol_pw":4.25,
         "avg_sol_pw":4.30,
         "power_usage":99.17,
         "avg_power_usage":99.57,
         "accepted_shares":3,
         "rejected_shares":2,
         "latency":292
      },
      {
         "gpu_id":5,
         "temperature":61,
         "sol_ps":416.42,
         "avg_sol_ps":424.63,
         "sol_pw":4.06,
         "avg_sol_pw":4.27,
         "power_usage":102.58,
         "avg_power_usage":99.46,
         "accepted_shares":4,
         "rejected_shares":1,
         "latency":290
      }
   ],
   "uptime":240,
   "contime":236,
   "server":"europe.equihash-hub.miningpoolhub.com",
   "port":17023,
   "user":"BugRoger.wupse",
   "version":"0.5.8",
   "error":null
}`
)

type MockedDSTMAPI struct {
	mock.Mock
}

func (m *MockedDSTMAPI) GetStat() (*getStat, error) {
	args := m.Called()

	var result getStat
	json.Unmarshal([]byte(args.String(0)), &result)

	return &result, args.Error(1)
}

func TestDSTMCollect(t *testing.T) {
	mockAPI := new(MockedDSTMAPI)

	mockAPI.On("GetStat").Return(GETSTAT, nil)

	miner := &DSTMClient{mockAPI}
	metrics, _ := miner.Collect()

	assert.Equal(t, "dstm", miner.Name())
	assert.Equal(t, "0.5.8", metrics.Version)
	assert.Equal(t, 240.0, metrics.Uptime)
	assert.Equal(t, "equihash", metrics.Algorithms[0].Name)
	assert.Equal(t, 15.0, metrics.Algorithms[0].Shares.Accepted)
	assert.Equal(t, 5.0, metrics.Algorithms[0].Shares.Rejected)
	assert.Equal(t, 0.0, metrics.Algorithms[0].Shares.Stale)
	assert.Equal(t, 422.49, metrics.Algorithms[0].Rates.ByGPU[0])
	assert.Equal(t, 427.96, metrics.Algorithms[0].Rates.ByGPU[1])
	assert.Equal(t, 431.08, metrics.Algorithms[0].Rates.ByGPU[2])
	assert.Equal(t, 424.16, metrics.Algorithms[0].Rates.ByGPU[3])
	assert.Equal(t, 421.73, metrics.Algorithms[0].Rates.ByGPU[4])
	assert.Equal(t, 416.42, metrics.Algorithms[0].Rates.ByGPU[5])
	assert.Equal(t, 2543.84, metrics.Algorithms[0].Rates.Total)
}
