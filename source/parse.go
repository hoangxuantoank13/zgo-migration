package source

import (
	"fmt"
	"regexp"
)

// ErrParse is a const error
var (
	ErrParse = fmt.Errorf("no match")
)

// DefaultParse, DefaultRegex is a const error
var (
	DefaultParse = Parse
	DefaultRegex = Regex
)

// Regex matches the following pattern:
//  123_name.up.ext
//  123_name.down.ext
var Regex = regexp.MustCompile(`^(.*)\.(` + string(Down) + `|` + string(Up) + `)\.(.*)$`)

// Parse returns Migration for matching Regex pattern.
func Parse(raw string) (*Migration, error) {
	m := Regex.FindStringSubmatch(raw)
	if len(m) == 4 {
		return &Migration{
			Identifier: m[1],
			Direction:  Direction(m[2]),
			Raw:        raw,
		}, nil
	}
	return nil, ErrParse
}
