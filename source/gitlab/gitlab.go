package gitlab

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	nurl "net/url"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
	"github.com/hoangxuantoank13/zgo-migration/source"
)

func init() {
	source.Register("gitlab", &Gitlab{})
}

//Error const
var (
	ErrNoUserInfo       = fmt.Errorf("no username:token provided")
	ErrNoAccessToken    = fmt.Errorf("no access token")
	ErrInvalidHost      = fmt.Errorf("invalid host")
	ErrInvalidProjectID = fmt.Errorf("invalid project id")
	ErrInvalidResponse  = fmt.Errorf("invalid response")
)

//Gitlab is struct of source
type Gitlab struct {
	client *gitlab.Client
	url    string

	projectID   string
	path        string
	listOptions *gitlab.ListTreeOptions
	getOptions  *gitlab.GetFileOptions
	migrations  *source.Migrations
}

//Config is config of source
type Config struct {
}

// Open is part of source.Driver interface implementation.
func (g *Gitlab) Open(url string) (source.Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.User == nil {
		return nil, ErrNoUserInfo
	}

	password, ok := u.User.Password()
	if !ok {
		return nil, ErrNoAccessToken
	}

	gn := &Gitlab{
		client:     gitlab.NewClient(nil, password),
		url:        url,
		migrations: source.NewMigrations(),
	}

	if u.Host != "" {
		uri := nurl.URL{
			Scheme: "https",
			Host:   u.Host,
		}

		err = gn.client.SetBaseURL(uri.String())
		if err != nil {
			return nil, ErrInvalidHost
		}
	}

	pe := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pe) < 1 {
		return nil, ErrInvalidProjectID
	}
	gn.projectID = pe[0]
	if len(pe) > 1 {
		gn.path = strings.Join(pe[1:], "/")
	}

	gn.listOptions = &gitlab.ListTreeOptions{
		Path: &gn.path,
		Ref:  &u.Fragment,
	}

	gn.getOptions = &gitlab.GetFileOptions{
		Ref: &u.Fragment,
	}

	if err := gn.readDirectory(); err != nil {
		return nil, err
	}

	return gn, nil
}

//WithInstance return instance of source
func WithInstance(client *gitlab.Client, config *Config) (source.Driver, error) {
	gn := &Gitlab{
		client:     client,
		migrations: source.NewMigrations(),
	}
	if err := gn.readDirectory(); err != nil {
		return nil, err
	}
	return gn, nil
}

func (g *Gitlab) readDirectory() error {
	nodes, response, err := g.client.Repositories.ListTree(g.projectID, g.listOptions)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return ErrInvalidResponse
	}

	for i := range nodes {
		m, err := g.nodeToMigration(nodes[i])
		if err != nil {
			continue
		}

		if !g.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", nodes[i].Name)
		}
	}

	return nil
}

func (g *Gitlab) nodeToMigration(node *gitlab.TreeNode) (*source.Migration, error) {
	m := source.Regex.FindStringSubmatch(node.Name)
	if len(m) == 4 {
		return &source.Migration{
			Identifier: m[1],
			Direction:  source.Direction(m[2]),
			Raw:        g.path + "/" + node.Name,
		}, nil
	}
	return nil, source.ErrParse
}

// Close is part of source.Driver interface implementation.
func (g *Gitlab) Close() error {
	return nil
}

// ReadUp is part of source.Driver interface implementation.
func (g *Gitlab) ReadUp(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := g.migrations.Up(identifier); ok {
		f, response, err := g.client.RepositoryFiles.GetFile(g.projectID, m.Raw, g.getOptions)
		if err != nil {
			return nil, "", err
		}

		if response.StatusCode != http.StatusOK {
			return nil, "", ErrInvalidResponse
		}

		content, err := base64.StdEncoding.DecodeString(f.Content)
		if err != nil {
			return nil, "", err
		}

		return ioutil.NopCloser(strings.NewReader(string(content))), m.Raw, nil
	}

	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: g.path, Err: os.ErrNotExist}
}

// ReadDown is part of source.Driver interface implementation.
func (g *Gitlab) ReadDown(identifier string) (r io.ReadCloser, raw string, err error) {
	if m, ok := g.migrations.Down(identifier); ok {
		f, response, err := g.client.RepositoryFiles.GetFile(g.projectID, m.Raw, g.getOptions)
		if err != nil {
			return nil, "", err
		}

		if response.StatusCode != http.StatusOK {
			return nil, "", ErrInvalidResponse
		}

		content, err := base64.StdEncoding.DecodeString(f.Content)
		if err != nil {
			return nil, "", err
		}

		return ioutil.NopCloser(strings.NewReader(string(content))), m.Raw, nil
	}

	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", identifier), Path: g.path, Err: os.ErrNotExist}
}

// GetAllSource is part of source.Driver interface implementation.
func (g *Gitlab) GetAllSource() (identifierSlice []string, err error) {
	return g.migrations.GetAllIdentifier()
}
