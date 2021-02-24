package tcpserver

import "regexp"

type RulesConfiguration struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Match      string `yaml:"match,omitempty"`
	matchRegex *regexp.Regexp
	Response   string `yaml:"response,omitempty"`
}

func NewRule(match string, response string) (*Rule, error) {
	regxp, err := regexp.Compile(match)
	if err != nil {
		return nil, err
	}

	return &Rule{Match: match, matchRegex: regxp, Response: response}, nil
}
