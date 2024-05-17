package resolvers_test

import (
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/dataRetriever/mock"
	"github.com/numbatx/gn-numbat/p2p"
)

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}
