package stub

import (
	"testing"

	"github.com/hoangxuantoank13/zgo-migration/source"
	st "github.com/hoangxuantoank13/zgo-migration/source/testing"
)

func Test(t *testing.T) {
	s := &Stub{}
	d, err := s.Open("")
	if err != nil {
		t.Fatal(err)
	}

	m := source.NewMigrations()
	m.Append(&source.Migration{Identifier: "a", Direction: source.Up})
	m.Append(&source.Migration{Identifier: "b", Direction: source.Down})
	m.Append(&source.Migration{Identifier: "c", Direction: source.Up})
	m.Append(&source.Migration{Identifier: "d", Direction: source.Up})
	m.Append(&source.Migration{Identifier: "e", Direction: source.Down})
	m.Append(&source.Migration{Identifier: "f", Direction: source.Down})
	m.Append(&source.Migration{Identifier: "g", Direction: source.Up})
	m.Append(&source.Migration{Identifier: "h", Direction: source.Down})

	d.(*Stub).Migrations = m

	st.Test(t, d)
}
