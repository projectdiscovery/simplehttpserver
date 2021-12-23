package tcpserver

import (
	"regexp"
	"strings"
)

// RulesConfiguration from yaml
type RulesConfiguration struct {
	Rules []Rule `yaml:"rules"`
}

// Rule to apply to various requests
type Rule struct {
	Name          string `yaml:"name,omitempty"`
	Match         string `yaml:"match,omitempty"`
	MatchContains string `yaml:"match-contains,omitempty"`
	matchRegex    *regexp.Regexp
	Response      string `yaml:"response,omitempty"`
}

// NewRule creates a new Rule - default is regex
func NewRule(match, response string) (*Rule, error) {
	return NewRegexRule(match, response)
}

// NewRegexRule returns a new regex-match Rule
func NewRegexRule(match, response string) (*Rule, error) {
	regxp, err := regexp.Compile(match)
	if err != nil {
		return nil, err
	}

	return &Rule{Match: match, matchRegex: regxp, Response: response}, nil
}

// NewLiteralRule returns a new literal-match Rule
func NewLiteralRule(match, response string) (*Rule, error) {
	return &Rule{MatchContains: match, Response: response}, nil
}

// NewRuleFromTemplate "copies" a new Rule
func NewRuleFromTemplate(r Rule) (newRule *Rule, err error) {
	newRule = &Rule{
		Name:          r.Name,
		Response:      r.Response,
		MatchContains: r.MatchContains,
		Match:         r.Match,
	}
	if newRule.Match != "" {
		newRule.matchRegex, err = regexp.Compile(newRule.Match)
	}

	return
}

// MatchInput returns if the input was matches with one of the matchers
func (r *Rule) MatchInput(input []byte) bool {
	if r.matchRegex != nil && r.matchRegex.Match(input) {
		return true
	} else if r.MatchContains != "" && strings.Contains(string(input), r.MatchContains) {
		return true
	}
	return false
}
