package tcpserver

import (
	"errors"
)

// BuildResponse according to rules
func (t *TCPServer) BuildResponse(data []byte) ([]byte, error) {
	// Process all the rules
	for _, rule := range t.options.rules {
		if rule.matchRegex.Match(data) {
			return []byte(rule.Response), nil
		}
	}
	return nil, errors.New("no matched rule")
}
