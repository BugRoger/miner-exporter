package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	SUMMARY = "NAME=ccminer;VER=2.2.4;API=1.9;ALGO=cryptonight;GPUS=6;KHS=0.61;SOLV=0;ACC=0;REJ=0;ACCMN=0.000;DIFF=120051.635035;NETKHS=0;POOLS=1;WAIT=5;UPTIME=90;TS=1518364735"

	THREADS = "GPU=0;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.31;KHW=0.00000;PLIM=0;ACC=0;REJ=0;HWF=0;I=9.9;THR=960|GPU=1;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.31;KHW=0.00000;PLIM=0;ACC=1;REJ=0;HWF=0;I=9.9;THR=960|GPU=2;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.27;KHW=0.00000;PLIM=0;ACC=0;REJ=0;HWF=0;I=9.9;THR=960|GPU=3;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.30;KHW=0.00000;PLIM=0;ACC=0;REJ=0;HWF=0;I=9.9;THR=960|GPU=4;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.30;KHW=0.00000;PLIM=0;ACC=2;REJ=0;HWF=0;I=9.9;THR=960|GPU=5;BUS=-1;CARD=GeForce GTX 1070;TEMP=0.0;POWER=0;FAN=0;RPM=0;FREQ=1683;MEMFREQ=4004;GPUF=0;MEMF=0;KHS=0.30;KHW=0.00000;PLIM=0;ACC=0;REJ=0;HWF=0;I=9.9;THR=960"

	POOL = "POOL=europe.cryptonight-hub.miningpoolhub.com:17024;ALGO=cryptonight;URL=stratum+tcp://europe.cryptonight-hub.miningpoolhub.com:17024;USER=bugroger.wupse;SOLV=0;ACC=3;REJ=0;STALE=1;H=0;JOB=;DIFF=84035.439844;BEST=662.573037;N2SZ=0;N2=;PING=412;DISCO=0;WAIT=5;UPTIME=0;LAST=45"
)

type MockedCCMinerAPI struct {
	mock.Mock
}

func (m *MockedCCMinerAPI) Summary() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockedCCMinerAPI) Threads() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockedCCMinerAPI) Pool() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestCollect(t *testing.T) {
	mockAPI := new(MockedCCMinerAPI)

	mockAPI.On("Summary").Return(SUMMARY, nil)
	mockAPI.On("Threads").Return(THREADS, nil)
	mockAPI.On("Pool").Return(POOL, nil)

	ccminer := &CCMinerClient{mockAPI}
	metrics, _ := ccminer.Collect()

	assert.Equal(t, "ccminer", ccminer.Name())
	assert.Equal(t, "2.2.4", metrics.Version)
	assert.Equal(t, 90.0, metrics.Uptime)
	assert.Equal(t, "cryptonight", metrics.Algorithms[0].Name)
	assert.Equal(t, 3.0, metrics.Algorithms[0].Shares.Accepted)
	assert.Equal(t, 0.0, metrics.Algorithms[0].Shares.Rejected)
	assert.Equal(t, 1.0, metrics.Algorithms[0].Shares.Stale)
	assert.Equal(t, 0.31, metrics.Algorithms[0].Rates.ByGPU[0])
	assert.Equal(t, 0.31, metrics.Algorithms[0].Rates.ByGPU[1])
	assert.Equal(t, 0.27, metrics.Algorithms[0].Rates.ByGPU[2])
	assert.Equal(t, 0.30, metrics.Algorithms[0].Rates.ByGPU[3])
	assert.Equal(t, 0.30, metrics.Algorithms[0].Rates.ByGPU[4])
	assert.Equal(t, 0.30, metrics.Algorithms[0].Rates.ByGPU[5])
	assert.Equal(t, 1.79, metrics.Algorithms[0].Rates.Total)
}
