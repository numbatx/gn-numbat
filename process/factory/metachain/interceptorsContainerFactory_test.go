package metachain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/factory"
	"github.com/numbatx/gn-numbat/process/factory/metachain"
	"github.com/numbatx/gn-numbat/process/mock"
	"github.com/numbatx/gn-numbat/storage"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
	return &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			if matchStrToErrOnCreate == "" {
				return nil
			}

			if strings.Contains(name, matchStrToErrOnCreate) {
				return errExpected
			}

			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			if matchStrToErrOnRegister == "" {
				return nil
			}

			if strings.Contains(topic, matchStrToErrOnRegister) {
				return errExpected
			}

			return nil
		},
	}
}

func createDataPools() dataRetriever.MetaPoolsHolder {
	pools := &mock.MetaPoolsHolderStub{
		MetaBlockNoncesCalled: func() dataRetriever.Uint64Cacher {
			return &mock.Uint64CacherStub{}
		},
		ShardHeadersCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		MiniBlockHashesCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		MetaChainBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}

	return pools
}

func createStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory

func TestNewInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		nil,
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		nil,
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

//------- Create

func TestInterceptorsContainerFactory_CreateTopicMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.MetachainBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.ShardHeadersForMetachainTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterForMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.MetachainBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.ShardHeadersForMetachainTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := metachain.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	icf, _ := metachain.NewInterceptorsContainerFactory(
		shardCoordinator,
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.ChronologyValidatorStub{},
		nil,
	)

	container, _ := icf.Create()

	numInterceptorsMetablock := 1
	numInterceptorsShardHeadersForMetachain := noOfShards
	totalInterceptors := numInterceptorsMetablock + numInterceptorsShardHeadersForMetachain

	assert.Equal(t, totalInterceptors, container.Len())
}
