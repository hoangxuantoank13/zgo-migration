package httpfs

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/spf13/viper"
)

const migrationConfigFile = "_migrations.json"

//const Error
var (
	ErrWrongFormat = errors.New("Wrong Fromat")
)

// PartialDriver is a helper service for creating new source drivers working with
// http.FileSystem instances. It implements all source.Driver interface methods
// except for Open(). New driver could embed this struct and add missing Open()
// method.
//
// To prepare PartialDriver for use Init() function.
type PartialDriver struct {
	migrations *source.Migrations
	fs         http.FileSystem
	path       string
}

// Init prepares not initialized PartialDriver instance to read migrations from a
// http.FileSystem instance and a relative path.
func (p *PartialDriver) Init(fs http.FileSystem, path string) error {
	mapFile, err := readMigrationConfig(path)
	if err != nil {
		return err
	}
	ms := source.NewMigrations()
	for _, files := range mapFile {
		for _, file := range files {
			m, err := source.DefaultParse(file)
			if err != nil {
				continue // ignore files that we can't parse
			}

			if !ms.Append(m) {
				return source.ErrDuplicateMigration{
					Migration: *m,
					FileName:  file,
				}
			}
		}
	}

	p.fs = fs
	p.path = path
	p.migrations = ms
	return nil
}

func readMigrationConfig(path string) (res []map[string]string, err error) {
	configPath := path + "/" + migrationConfigFile
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	migrations := viper.Get("migrations")
	m, ok := migrations.([]interface{})

	if !ok {
		return nil, ErrWrongFormat
	}
	ret := make([]map[string]string, len(m))
	for i := 0; i < len(m); i++ {
		ma, ok := m[i].(map[string]interface{})
		if !ok {
			return nil, ErrWrongFormat
		}
		it := make(map[string]string)
		for k, v := range ma {
			v, ok := v.(string)
			if !ok {
				return nil, ErrWrongFormat
			}
			it[k] = v
		}
		ret[i] = it
	}
	return ret, nil
}

// Close is part of source.Driver interface implementation. This is a no-op.
func (p *PartialDriver) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (p *PartialDriver) ReadUp(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := p.migrations.Up(identifier); ok {
		body, err := p.fs.Open(path.Join(p.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return body, m.Raw, nil
	}
	return nil, "", &os.PathError{
		Op:   "read up for identifier " + identifier,
		Path: p.path,
		Err:  os.ErrNotExist,
	}
}

// ReadDown is part of source.Driver interface implementation.
func (p *PartialDriver) ReadDown(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := p.migrations.Down(identifier); ok {
		body, err := p.fs.Open(path.Join(p.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return body, m.Raw, nil
	}
	return nil, "", &os.PathError{
		Op:   "read down for identifier " + identifier,
		Path: p.path,
		Err:  os.ErrNotExist,
	}
}

// GetAllSource is part of source.Driver interface implementation.
func (p *PartialDriver) GetAllSource() (identifierSlice []string, err error) {
	return p.migrations.GetAllIdentifier()
}
