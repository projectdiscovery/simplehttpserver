package tcpserver

import "regexp"

// RulesConfiguration from yaml
type RulesConfiguration struct {
	Rules []Rule `yaml:"rules"`
}

// Rule to apply to various requests
type Rule struct {
	Match      string `yaml:"match,omitempty"`
	matchRegex *regexp.Regexp
	Response   string `yaml:"response,omitempty"`
}

// NewRule from model
func NewRule(match, response string) (*Rule, error) {
	regxp, err := regexp.Compile(match)
	if err != nil {
		return nil, err
	}

	return &Rule{Match: match, matchRegex: regxp, Response: response}, nil
}
