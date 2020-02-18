package awss3

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/golang-migrate/migrate/v4/source"
)

func init() {
	source.Register("s3", &s3Driver{})
}

//s3Driver is struct of source
type s3Driver struct {
	s3client   s3iface.S3API
	config     *Config
	migrations *source.Migrations
}

//Config is config of source
type Config struct {
	Bucket string
	Prefix string
}

// Open is part of source.Driver interface implementation.
func (s *s3Driver) Open(folder string) (source.Driver, error) {
	config, err := parseURI(folder)
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return WithInstance(s3.New(sess), config)
}

//WithInstance return instance of source
func WithInstance(s3client s3iface.S3API, config *Config) (source.Driver, error) {
	driver := &s3Driver{
		config:     config,
		s3client:   s3client,
		migrations: source.NewMigrations(),
	}

	if err := driver.loadMigrations(); err != nil {
		return nil, err
	}

	return driver, nil
}

func parseURI(uri string) (*Config, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	prefix := strings.Trim(u.Path, "/")
	if prefix != "" {
		prefix += "/"
	}

	return &Config{
		Bucket: u.Host,
		Prefix: prefix,
	}, nil
}

func (s *s3Driver) loadMigrations() error {
	output, err := s.s3client.ListObjects(&s3.ListObjectsInput{
		Bucket:    aws.String(s.config.Bucket),
		Prefix:    aws.String(s.config.Prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return err
	}
	for _, object := range output.Contents {
		_, fileName := path.Split(aws.StringValue(object.Key))
		m, err := source.DefaultParse(fileName)
		if err != nil {
			continue
		}
		if !s.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", aws.StringValue(object.Key))
		}
	}
	return nil
}

// Close is part of source.Driver interface implementation.
func (s *s3Driver) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (s *s3Driver) ReadUp(identifier string) (io.ReadCloser, string, error) {
	if m, ok := s.migrations.Up(identifier); ok {
		return s.open(m)
	}
	return nil, "", os.ErrNotExist
}

// ReadDown is part of source.Driver interface implementation.
func (s *s3Driver) ReadDown(identifier string) (io.ReadCloser, string, error) {
	if m, ok := s.migrations.Down(identifier); ok {
		return s.open(m)
	}
	return nil, "", os.ErrNotExist
}

func (s *s3Driver) open(m *source.Migration) (io.ReadCloser, string, error) {
	key := path.Join(s.config.Prefix, m.Raw)
	object, err := s.s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}
	return object.Body, m.Raw, nil
}

// GetAllSource is part of source.Driver interface implementation.
func (s *s3Driver) GetAllSource() (identifierSlice []string, err error) {
	return s.migrations.GetAllIdentifier()
}