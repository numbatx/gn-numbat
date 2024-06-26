package spos_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/numbatx/gn-numbat/consensus/mock"
	"github.com/numbatx/gn-numbat/consensus/spos"
	"github.com/numbatx/gn-numbat/consensus/spos/bls"
	"github.com/stretchr/testify/assert"
)

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, fmt.Sprintf("%c",i+65))

	}
	return eligibleList
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)
	indexLeader := 1
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	pFTThreshold := consensusGroupSize*2/3 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, pFTThreshold)
	rthr.SetThreshold(3, pFTThreshold)
	rthr.SetThreshold(4, pFTThreshold)
	rthr.SetThreshold(5, pFTThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.RoundIndex = 0
	return cns
}

func TestSubround_NewSubroundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		nil,
		ch,
		container,
	)

	assert.Equal(t, spos.ErrNilConsensusState, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilChannelShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		nil,
		container,
	)

	assert.Equal(t, spos.ErrNilChannel, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		nil,
	)

	assert.Equal(t, spos.ErrNilConsensusCore, err)
	assert.Nil(t, sr)
}

func TestSubround_NilContainerBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetBlockchain(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubround_NilContainerBlockprocessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetBlockProcessor(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubround_NilContainerBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetBootStrapper(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlootstraper, err)
}

func TestSubround_NilContainerChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetChronology(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestSubround_NilContainerHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetHasher(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubround_NilContainerMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetMarshalizer(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubround_NilContainerMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetMultiSigner(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubround_NilContainerRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetRounder(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubround_NilContainerShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetShardCoordinator(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubround_NilContainerSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetSyncTimer(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubround_NilContainerValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(nil)

	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilValidatorGroupSelector, err)
}

func TestSubround_NewSubroundShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	sr, err := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	assert.Nil(t, err)

	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.NotNil(t, sr)
}

func TestSubround_DoWorkShouldReturnFalseWhenJobFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.Job = nil
	sr.Check = func() bool {
		return true
	}

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)

	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenCheckFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = nil

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenConsensusIsNotDone(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(-1),
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobAndConsensusAreDone(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return true
	}

	maxTime := time.Now().Add(100 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	r := sr.DoWork(rounderMock)
	assert.True(t, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobIsDoneAndConsensusIsDoneAfterAWhile(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		container,
	)

	var mut sync.RWMutex
	mut.Lock()
	checkSuccess := false
	mut.Unlock()

	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		mut.RLock()
		defer mut.RUnlock()
		return checkSuccess
	}

	maxTime := time.Now().Add(2000 * time.Millisecond)
	rounderMock := &mock.RounderMock{}
	rounderMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return maxTime.Sub(time.Now())
	}

	go func() {
		time.Sleep(1000 * time.Millisecond)

		mut.Lock()
		checkSuccess = true
		mut.Unlock()

		ch <- true
	}()

	r := sr.DoWork(rounderMock)

	assert.True(t, r)
}

func TestSubround_Previous(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int(bls.SrSignature),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int(bls.SrStartRound), sr.Previous())
}

func TestSubround_Current(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int(bls.SrSignature),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int(bls.SrBlock), sr.Current())
}

func TestSubround_Next(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int(bls.SrSignature),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int(bls.SrSignature), sr.Next())
}

func TestSubround_StartTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetRounder(initRounderMock())
	sr, _ := spos.NewSubround(
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(25*roundTimeDuration/100),
		int64(40*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.StartTime())
}

func TestSubround_EndTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	container.SetRounder(initRounderMock())
	sr, _ := spos.NewSubround(
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int(bls.SrSignature),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.EndTime())
}

func TestSubround_Name(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		int(bls.SrStartRound),
		int(bls.SrBlock),
		int(bls.SrSignature),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	sr.Job = func() bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, "(BLOCK)", sr.Name())
}
