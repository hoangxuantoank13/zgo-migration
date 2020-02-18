// Package testing has the source tests.
// All source drivers must pass the Test function.
// This lives in it's own package so it stays a test dependency.
package testing

import (
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4/source"
)

// Test runs tests against source implementations.
// It assumes that the driver tests has access to the following migrations:
//
// u = up migration, d = down migration, n = version
//  |  1  |  -  |  3  |  4  |  5  |  -  |  7  |
//  | u d |  -  | u   | u d |   d |  -  | u d |
//
// See source/stub/stub_test.go or source/file/file_test.go for an example.
func Test(t *testing.T, d source.Driver) {
	TestReadUp(t, d)
	TestReadDown(t, d)
}

//TestReadUp is a testing
func TestReadUp(t *testing.T, d source.Driver) {
	tt := []struct {
		identifier string
		expectErr  error
		expectUp   bool
	}{
		{identifier: "0", expectErr: os.ErrNotExist},
		{identifier: "1", expectErr: nil, expectUp: true},
		{identifier: "2", expectErr: os.ErrNotExist},
		{identifier: "3", expectErr: nil, expectUp: true},
		{identifier: "4", expectErr: nil, expectUp: true},
		{identifier: "5", expectErr: os.ErrNotExist},
		{identifier: "6", expectErr: os.ErrNotExist},
		{identifier: "7", expectErr: nil, expectUp: true},
		{identifier: "8", expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		up, identifier, err := d.ReadUp(v.identifier)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			if len(identifier) == 0 {
				t.Errorf("expected identifier not to be empty, in %v", i)
			}

			if v.expectUp && up == nil {
				t.Errorf("expected up not to be nil, in %v", i)
			} else if !v.expectUp && up != nil {
				t.Errorf("expected up to be nil, got %v, in %v", up, i)
			}
		}
	}
}

//TestReadDown is a testing
func TestReadDown(t *testing.T, d source.Driver) {
	tt := []struct {
		identifier string
		expectErr  error
		expectDown bool
	}{
		{identifier: "0", expectErr: os.ErrNotExist},
		{identifier: "1", expectErr: nil, expectDown: true},
		{identifier: "2", expectErr: os.ErrNotExist},
		{identifier: "3", expectErr: os.ErrNotExist},
		{identifier: "4", expectErr: nil, expectDown: true},
		{identifier: "5", expectErr: nil, expectDown: true},
		{identifier: "6", expectErr: os.ErrNotExist},
		{identifier: "7", expectErr: nil, expectDown: true},
		{identifier: "8", expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		down, identifier, err := d.ReadDown(v.identifier)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			if len(identifier) == 0 {
				t.Errorf("expected identifier not to be empty, in %v", i)
			}

			if v.expectDown && down == nil {
				t.Errorf("expected down not to be nil, in %v", i)
			} else if !v.expectDown && down != nil {
				t.Errorf("expected down to be nil, got %v, in %v", down, i)
			}
		}
	}
}
