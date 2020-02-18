package stub

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hoangxuantoank13/zgo-migration/source"
)

func init() {
	source.Register("stub", &Stub{})
}

//Config is config of source
type Config struct{}

// d, _ := source.Open("stub://")
// d.(*stub.Stub).Migrations =

//Stub is source
type Stub struct {
	URL        string
	Instance   interface{}
	Migrations *source.Migrations
	Config     *Config
}

// Open returns a a new driver instance configured with parameters
// coming from the URL string. Migrate will call this function
// only once per instance.
func (s *Stub) Open(url string) (source.Driver, error) {
	return &Stub{
		URL:        url,
		Migrations: source.NewMigrations(),
		Config:     &Config{},
	}, nil
}

//WithInstance return Stub instance
func WithInstance(instance interface{}, config *Config) (source.Driver, error) {
	return &Stub{
		Instance:   instance,
		Migrations: source.NewMigrations(),
		Config:     config,
	}, nil
}

// Close closes the underlying source instance managed by the driver.
// Migrate will call this function only once per instance.
func (s *Stub) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (s *Stub) ReadUp(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := s.Migrations.Up(identifier); ok {
		return ioutil.NopCloser(bytes.NewBufferString(m.Identifier)), fmt.Sprintf("%v.up.stub", identifier), nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read up version %v", identifier), Path: s.URL, Err: os.ErrNotExist}
}

// ReadDown is part of source.Driver interface implementation.
func (s *Stub) ReadDown(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := s.Migrations.Down(identifier); ok {
		return ioutil.NopCloser(bytes.NewBufferString(m.Identifier)), fmt.Sprintf("%v.down.stub", identifier), nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read down version %v", identifier), Path: s.URL, Err: os.ErrNotExist}
}

// GetAllSource is part of source.Driver interface implementation.
func (s *Stub) GetAllSource() (identifierSlice []string, err error) {
	return s.Migrations.GetAllIdentifier()
}
