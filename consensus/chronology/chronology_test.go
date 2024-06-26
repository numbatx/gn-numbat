package chronology_test

import (
	"testing"
	"time"

	"github.com/numbatx/gn-numbat/consensus"
	"github.com/numbatx/gn-numbat/consensus/chronology"
	"github.com/numbatx/gn-numbat/consensus/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundHandlerMock() *mock.SubroundHandlerMock {
	srm := &mock.SubroundHandlerMock{}
	srm.CurrentCalled = func() int {
		return 0
	}
	srm.NextCalled = func() int {
		return 1
	}
	srm.DoWorkCalled = func(rounder consensus.Rounder) bool {
		return false
	}
	srm.NameCalled = func() string {
		return "(TEST)"
	}
	return srm
}

func TestChronology_NewChronologyNilRounderShouldFail(t *testing.T) {
	t.Parallel()
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, err := chronology.NewChronology(
		genesisTime,
		nil,
		syncTimerMock)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilRounder)
}

func TestChronology_NewChronologyNilSyncerShouldFail(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	genesisTime := time.Now()
	chr, err := chronology.NewChronology(
		genesisTime,
		rounderMock,
		nil)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilSyncTimer)
}

func TestChronology_NewChronologyShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, err := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	assert.NotNil(t, chr)
	assert.Nil(t, err)
}

func TestChronology_AddSubroundShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))
}

func TestChronology_RemoveAllSubroundsShouldReturnEmptySubroundHandlersArray(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))
	chr.RemoveAllSubrounds()
	assert.Equal(t, 0, len(chr.SubroundHandlers()))
}

func TestChronology_StartRoundShouldReturnWhenRoundIndexIsNegative(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	rounderMock.IndexCalled = func() int32 {
		return -1
	}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenLoadSubroundHandlerReturnsNil(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	initSubroundHandlerMock()
	chr.StartRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenDoWorkReturnsFalse(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	rounderMock.UpdateRound(rounderMock.TimeStamp(), rounderMock.TimeStamp().Add(rounderMock.TimeDuration()))
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_StartRoundShouldWork(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	rounderMock.UpdateRound(rounderMock.TimeStamp(), rounderMock.TimeStamp().Add(rounderMock.TimeDuration()))
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()
	srm.DoWorkCalled = func(rounder consensus.Rounder) bool {
		return true
	}
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, srm.Next(), chr.SubroundId())
}

func TestChronology_UpdateRoundShouldInitRound(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()
	chr.AddSubround(srm)
	chr.UpdateRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_LoadSubrounderShouldReturnNilWhenSubroundHandlerNotExists(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_LoadSubrounderShouldReturnNilWhenIndexIsOutOfBound(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	genesisTime := time.Now()
	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	chr.SetSubroundHandlers(make([]consensus.SubroundHandler, 0))

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_InitRoundShouldNotSetSubroundWhenRoundIndexIsNegative(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	chr, _ := chronology.NewChronology(
		syncTimerMock.CurrentTime(),
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	rounderMock.IndexCalled = func() int32 {
		return -1
	}
	chr.InitRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_InitRoundShouldSetSubroundWhenRoundIndexIsPositive(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	rounderMock.UpdateRound(rounderMock.TimeStamp(), rounderMock.TimeStamp().Add(rounderMock.TimeDuration()))
	syncTimerMock := &mock.SyncTimerMock{}
	chr, _ := chronology.NewChronology(
		syncTimerMock.CurrentTime(),
		rounderMock,
		syncTimerMock)

	sr := initSubroundHandlerMock()
	chr.AddSubround(sr)
	chr.InitRound()

	assert.Equal(t, sr.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldNotUpdateRoundWhenCurrentRoundIsNotFinished(t *testing.T) {
	t.Parallel()
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	chr, _ := chronology.NewChronology(
		syncTimerMock.CurrentTime(),
		rounderMock,
		syncTimerMock)

	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, int32(0), rounderMock.Index())
}

func TestChronology_StartRoundShouldUpdateRoundWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	chr, _ := chronology.NewChronology(
		syncTimerMock.CurrentTime(),
		rounderMock,
		syncTimerMock)

	chr.SetSubroundId(-1)
	chr.StartRound()

	assert.Equal(t, int32(1), rounderMock.Index())
}
