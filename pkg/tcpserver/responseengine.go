package tcpserver

import (
	"errors"
)

// BuildResponse according to rules
func (t *TCPServer) BuildResponse(data []byte) ([]byte, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	// Process all the rules
	for _, rule := range t.rules {
		if rule.MatchInput(data) {
			return []byte(rule.Response), nil
		}
	}
	return nil, errors.New("no matched rule")
}
