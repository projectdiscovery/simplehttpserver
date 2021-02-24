package tcpserver

import (
	"errors"
)

func (t *TCPServer) BuildResponse(data []byte) ([]byte, error) {
	// Process all the rules
	for _, rule := range t.options.rules {
		if rule.matchRegex.Match(data) {
			return []byte(rule.Response), nil
		}
	}
	return nil, errors.New("No matched rule")
}
