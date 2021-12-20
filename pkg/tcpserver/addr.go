package tcpserver

// ContextType is the key type stored in ctx
type ContextType string

var (
	// Addr is the contextKey where the net.Addr is stored
	Addr ContextType = "addr"
)
