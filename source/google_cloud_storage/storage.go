package googlecloudstorage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/golang-migrate/migrate/v4/source"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func init() {
	source.Register("gcs", &gcs{})
}

//gcs is a struct of source
type gcs struct {
	bucket     *storage.BucketHandle
	prefix     string
	migrations *source.Migrations
}

// Open is part of source.Driver interface implementation.
func (g *gcs) Open(folder string) (source.Driver, error) {
	u, err := url.Parse(folder)
	if err != nil {
		return nil, err
	}
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}
	driver := gcs{
		bucket:     client.Bucket(u.Host),
		prefix:     strings.Trim(u.Path, "/") + "/",
		migrations: source.NewMigrations(),
	}
	err = driver.loadMigrations()
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (g *gcs) loadMigrations() error {
	iter := g.bucket.Objects(context.Background(), &storage.Query{
		Prefix:    g.prefix,
		Delimiter: "/",
	})
	object, err := iter.Next()
	for ; err == nil; object, err = iter.Next() {
		_, fileName := path.Split(object.Name)
		m, parseErr := source.DefaultParse(fileName)
		if parseErr != nil {
			continue
		}
		if !g.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", object.Name)
		}
	}
	if err != iterator.Done {
		return err
	}
	return nil
}

// Close is part of source.Driver interface implementation.
func (g *gcs) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (g *gcs) ReadUp(identifier string) (io.ReadCloser, string, error) {
	if m, ok := g.migrations.Up(identifier); ok {
		return g.open(m)
	}
	return nil, "", os.ErrNotExist
}

// ReadDown is part of source.Driver interface implementation.
func (g *gcs) ReadDown(identifier string) (io.ReadCloser, string, error) {
	if m, ok := g.migrations.Down(identifier); ok {
		return g.open(m)
	}
	return nil, "", os.ErrNotExist
}

func (g *gcs) open(m *source.Migration) (io.ReadCloser, string, error) {
	objectPath := path.Join(g.prefix, m.Raw)
	reader, err := g.bucket.Object(objectPath).NewReader(context.Background())
	if err != nil {
		return nil, "", err
	}
	return reader, m.Raw, nil
}

// GetAllSource is part of source.Driver interface implementation.
func (g *gcs) GetAllSource() (identifierSlice []string, err error) {
	return g.migrations.GetAllIdentifier()
}