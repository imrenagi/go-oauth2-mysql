package mysql

// ClientStoreOption is the configuration options type for client store
type ClientStoreOption func(s *ClientStore)

// WithClientStoreTableName returns option that sets client store table name
func WithClientStoreTableName(tableName string) ClientStoreOption {
	return func(s *ClientStore) {
		s.tableName = tableName
	}
}

// WithClientStoreInitTableDisabled returns option that disables table creation on client store instantiation
func WithClientStoreInitTableDisabled() ClientStoreOption {
	return func(s *ClientStore) {
		s.initTableDisabled = true
	}
}
