package source

import (
	"sort"
)

// Direction is either up or down.
type Direction string

//Direction const
const (
	Down Direction = "down"
	Up   Direction = "up"
)

// Migration is a helper struct for source drivers that need to
// build the full directory tree in memory.
// Migration is fully independent from migrate.Migration.
type Migration struct {
	// Identifier can be any string that helps identifying
	// this migration in the source.
	Identifier string

	// Direction is either Up or Down.
	Direction Direction

	// Raw holds the raw location path to this migration in source.
	// ReadUp and ReadDown will use this.
	Raw string
}

// Migrations wraps Migration and has an internal index
// to keep track of Migration order.
type Migrations struct {
	//index      stringSlice
	migrations map[string]map[Direction]*Migration
}

//NewMigrations constructor
func NewMigrations() *Migrations {
	return &Migrations{
		//index:      make(stringSlice, 0),
		migrations: make(map[string]map[Direction]*Migration),
	}
}

//Append append a Migration to Migrations
func (i *Migrations) Append(m *Migration) (ok bool) {
	if m == nil {
		return false
	}

	if i.migrations[m.Identifier] == nil {
		i.migrations[m.Identifier] = make(map[Direction]*Migration)
	}

	// reject duplicate versions
	if _, dup := i.migrations[m.Identifier][m.Direction]; dup {
		return false
	}

	i.migrations[m.Identifier][m.Direction] = m
	//i.buildIndex()

	return true
}

//Up get the up migration file
func (i *Migrations) Up(identifier string) (m *Migration, ok bool) {
	if _, ok := i.migrations[identifier]; ok {
		if mx, ok := i.migrations[identifier][Up]; ok {
			return mx, true
		}
	}
	return nil, false
}

//Down get the down migration file
func (i *Migrations) Down(identifier string) (m *Migration, ok bool) {
	if _, ok := i.migrations[identifier]; ok {
		if mx, ok := i.migrations[identifier][Down]; ok {
			return mx, true
		}
	}
	return nil, false
}

//GetAllIdentifier get all Identifier
func (i *Migrations) GetAllIdentifier() (IdentifierSlice []string, err error) {
	keys := make([]string, 0, len(i.migrations))
    for k := range i.migrations {
        keys = append(keys, k)
    }
	return keys, nil
}

type stringSlice []string

func (s stringSlice) Len() int {
	return len(s)
}

func (s stringSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s stringSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s stringSlice) Search(x string) int {
	return sort.Search(len(s), func(i int) bool { return s[i] >= x })
}
