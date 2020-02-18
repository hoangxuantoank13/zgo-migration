package github

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"os"
	"path"
	"strings"
)

import (
	"github.com/google/go-github/github"
	"github.com/hoangxuantoank13/zgo-migration/source"
)

func init() {
	source.Register("github", &Github{})
}

//Error const
var (
	ErrNoUserInfo          = fmt.Errorf("no username:token provided")
	ErrNoAccessToken       = fmt.Errorf("no access token")
	ErrInvalidRepo         = fmt.Errorf("invalid repo")
	ErrInvalidGithubClient = fmt.Errorf("expected *github.Client")
	ErrNoDir               = fmt.Errorf("no directory")
)

//Github is struct of source
type Github struct {
	config     *Config
	client     *github.Client
	options    *github.RepositoryContentGetOptions
	migrations *source.Migrations
}

//Config is config of source
type Config struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

// Open is part of source.Driver interface implementation.
func (g *Github) Open(url string) (source.Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.User == nil {
		return nil, ErrNoUserInfo
	}

	password, ok := u.User.Password()
	if !ok {
		return nil, ErrNoUserInfo
	}

	tr := &github.BasicAuthTransport{
		Username: u.User.Username(),
		Password: password,
	}

	gn := &Github{
		client:     github.NewClient(tr.Client()),
		migrations: source.NewMigrations(),
		options:    &github.RepositoryContentGetOptions{Ref: u.Fragment},
	}

	gn.ensureFields()

	// set owner, repo and path in repo
	gn.config.Owner = u.Host
	pe := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pe) < 1 {
		return nil, ErrInvalidRepo
	}
	gn.config.Repo = pe[0]
	if len(pe) > 1 {
		gn.config.Path = strings.Join(pe[1:], "/")
	}

	if err := gn.readDirectory(); err != nil {
		return nil, err
	}

	return gn, nil
}

//WithInstance return instance of source
func WithInstance(client *github.Client, config *Config) (source.Driver, error) {
	gn := &Github{
		client:     client,
		config:     config,
		migrations: source.NewMigrations(),
		options:    &github.RepositoryContentGetOptions{Ref: config.Ref},
	}

	if err := gn.readDirectory(); err != nil {
		return nil, err
	}

	return gn, nil
}

func (g *Github) readDirectory() error {
	g.ensureFields()

	fileContent, dirContents, _, err := g.client.Repositories.GetContents(
		context.Background(),
		g.config.Owner,
		g.config.Repo,
		g.config.Path,
		g.options,
	)

	if err != nil {
		return err
	}
	if fileContent != nil {
		return ErrNoDir
	}

	for _, fi := range dirContents {
		m, err := source.DefaultParse(*fi.Name)
		if err != nil {
			continue // ignore files that we can't parse
		}
		if !g.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", *fi.Name)
		}
	}

	return nil
}

func (g *Github) ensureFields() {
	if g.config == nil {
		g.config = &Config{}
	}
}

// Close is part of source.Driver interface implementation.
func (g *Github) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (g *Github) ReadUp(identifier string) (r io.ReadCloser, raw string, err error) {
	g.ensureFields()

	if m, ok := g.migrations.Up(identifier); ok {
		file, _, _, err := g.client.Repositories.GetContents(
			context.Background(),
			g.config.Owner,
			g.config.Repo,
			path.Join(g.config.Path, m.Raw),
			g.options,
		)

		if err != nil {
			return nil, "", err
		}
		if file != nil {
			r, err := file.GetContent()
			if err != nil {
				return nil, "", err
			}
			return ioutil.NopCloser(strings.NewReader(r)), m.Raw, nil
		}
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: g.config.Path, Err: os.ErrNotExist}
}

// ReadDown is part of source.Driver interface implementation.
func (g *Github) ReadDown(identifier string) (r io.ReadCloser, raw string, err error) {
	g.ensureFields()

	if m, ok := g.migrations.Down(identifier); ok {
		file, _, _, err := g.client.Repositories.GetContents(
			context.Background(),
			g.config.Owner,
			g.config.Repo,
			path.Join(g.config.Path, m.Raw),
			g.options,
		)

		if err != nil {
			return nil, "", err
		}
		if file != nil {
			r, err := file.GetContent()
			if err != nil {
				return nil, "", err
			}
			return ioutil.NopCloser(strings.NewReader(r)), m.Raw, nil
		}
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: g.config.Path, Err: os.ErrNotExist}
}

// GetAllSource is part of source.Driver interface implementation.
func (g *Github) GetAllSource() (identifierSlice []string, err error) {
	return g.migrations.GetAllIdentifier()
}
