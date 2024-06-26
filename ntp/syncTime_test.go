package ntp_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/beevik/ntp"
	ntp2 "github.com/numbatx/gn-numbat/ntp"
	"github.com/stretchr/testify/assert"
)

var responseMock1 *ntp.Response
var failNtpMock1 = false
var responseMock2 *ntp.Response
var failNtpMock2 = false
var responseMock3 *ntp.Response
var failNtpMock3 = false

var errNtpMock = errors.New("NTP Mock generic error")
var queryMock4Call = 0
var mutex = sync.Mutex{}

func queryMock1(host string) (*ntp.Response, error) {
	fmt.Printf("Host: %s\n", host)

	if failNtpMock1 {
		return nil, errNtpMock
	}

	return responseMock1, nil
}

func queryMock2(host string) (*ntp.Response, error) {
	fmt.Printf("Host: %s\n", host)

	if failNtpMock2 {
		return nil, errNtpMock
	}

	return responseMock2, nil
}

func queryMock3(host string) (*ntp.Response, error) {
	fmt.Printf("Host: %s\n", host)

	if failNtpMock3 {
		return nil, errNtpMock
	}

	return responseMock3, nil
}

func queryMock4(host string) (*ntp.Response, error) {
	fmt.Printf("Host: %s\n", host)

	mutex.Lock()
	queryMock4Call++
	mutex.Unlock()

	return nil, errNtpMock
}

func TestHandleErrorInDoSync(t *testing.T) {
	failNtpMock1 = true
	st := ntp2.NewSyncTime(time.Millisecond, queryMock1)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)

	st.SetClockOffset(1234)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Duration(1234))

}

func TestValueInDoSync(t *testing.T) {
	responseMock2 = &ntp.Response{ClockOffset: 23456}

	failNtpMock2 = false
	st := ntp2.NewSyncTime(time.Millisecond, queryMock2)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.Sync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)

	st.SetClockOffset(1234)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestGetOffset(t *testing.T) {
	responseMock3 = &ntp.Response{ClockOffset: 23456}

	failNtpMock3 = false
	st := ntp2.NewSyncTime(time.Millisecond, queryMock3)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.Sync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestCallQuery(t *testing.T) {
	st := ntp2.NewSyncTime(time.Millisecond, queryMock4)
	go st.StartSync()

	assert.NotNil(t, st.Query())
	assert.Equal(t, time.Millisecond, st.SyncPeriod())

	// wait a few cycles
	time.Sleep(time.Millisecond * 100)

	mutex.Lock()
	qmc := queryMock4Call
	mutex.Unlock()
	assert.NotEqual(t, qmc, 0)

	fmt.Printf("Current time: %v\n", st.FormattedCurrentTime())
}
