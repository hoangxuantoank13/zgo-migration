package bindata

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hoangxuantoank13/zgo-migration/source"
)

//AssetFunc is a a function
type AssetFunc func(name string) ([]byte, error)

//Resource is a struct
func Resource(names []string, afn AssetFunc) *AssetSource {
	return &AssetSource{
		Names:     names,
		AssetFunc: afn,
	}
}

//AssetSource is a const
type AssetSource struct {
	Names     []string
	AssetFunc AssetFunc
}

func init() {
	source.Register("go-bindata", &Bindata{})
}

//Bindata is a struct of source
type Bindata struct {
	path        string
	assetSource *AssetSource
	migrations  *source.Migrations
}

// Open is part of source.Driver interface implementation.
func (b *Bindata) Open(url string) (source.Driver, error) {
	return nil, fmt.Errorf("not yet implemented")
}

//ErrNoAssetSource is a const Error
var (
	ErrNoAssetSource = fmt.Errorf("expects *AssetSource")
)

//WithInstance return instance of source
func WithInstance(instance interface{}) (source.Driver, error) {
	if _, ok := instance.(*AssetSource); !ok {
		return nil, ErrNoAssetSource
	}
	as := instance.(*AssetSource)

	bn := &Bindata{
		path:        "<go-bindata>",
		assetSource: as,
		migrations:  source.NewMigrations(),
	}

	for _, fi := range as.Names {
		m, err := source.DefaultParse(fi)
		if err != nil {
			continue // ignore files that we can't parse
		}

		if !bn.migrations.Append(m) {
			return nil, fmt.Errorf("unable to parse file %v", fi)
		}
	}

	return bn, nil
}

// Close is part of source.Driver interface implementation.
func (b *Bindata) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (b *Bindata) ReadUp(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := b.migrations.Up(identifier); ok {
		body, err := b.assetSource.AssetFunc(m.Raw)
		if err != nil {
			return nil, "", err
		}
		return ioutil.NopCloser(bytes.NewReader(body)), m.Raw, nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: b.path, Err: os.ErrNotExist}
}

// ReadDown is part of source.Driver interface implementation.
func (b *Bindata) ReadDown(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := b.migrations.Down(identifier); ok {
		body, err := b.assetSource.AssetFunc(m.Raw)
		if err != nil {
			return nil, "", err
		}
		return ioutil.NopCloser(bytes.NewReader(body)), m.Raw, nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: b.path, Err: os.ErrNotExist}
}

// GetAllSource is part of source.Driver interface implementation.
func (b *Bindata) GetAllSource() (identifierSlice []string, err error) {
	return b.migrations.GetAllIdentifier()
}
