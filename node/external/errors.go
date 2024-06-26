package external

import "github.com/pkg/errors"

// ErrOperationNotSupported signals that the operation is not supported
var ErrOperationNotSupported = errors.New("operation is not supported on this node")

// ErrInvalidValue signals that the value provided is invalid
var ErrInvalidValue = errors.New("provided value is invalid")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrMetablockNotFoundInLocalStorage signals that the required metablock was not found in local storage
var ErrMetablockNotFoundInLocalStorage = errors.New("metablock was not found in local storage")

// ErrNilBlockChain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilStore signals that the provided storage service is nil
var ErrNilStore = errors.New("nil data storage service")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilProposerResolver is raised when a valid proposer resolver is expected but nil used
var ErrNilProposerResolver = errors.New("proposer validator is nil")

// ErrNilHash signals that a nil hash has been provided
var ErrNilHash = errors.New("nil hash")
