package factory

const (
	// TransactionTopic is the topic used for sharing transactions
	TransactionTopic = "transactions"
	// HeadersTopic is the topic used for sharing block headers
	HeadersTopic = "headers"
	// MiniBlocksTopic is the topic used for sharing mini blocks
	MiniBlocksTopic = "txBlockBodies"
	// PeerChBodyTopic is used for sharing peer change block bodies
	PeerChBodyTopic = "peerChangeBlockBodies"
	// MetachainBlocksTopic is used for sharing metachain block headers between shards
	MetachainBlocksTopic = "metachainBlocks"
	// ShardHeadersForMetachainTopic is used for sharing shards block headers to the metachain nodes
	ShardHeadersForMetachainTopic = "shardHeadersForMetachain"
)
