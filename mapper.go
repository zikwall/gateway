package gateway

import (
	"fmt"
	"net/url"
	"os"

	"github.com/armon/go-radix"
	"gopkg.in/yaml.v3"
)

type Mapper struct {
	services []*description
	def      *description
	routers  *radix.Tree
}

func (m *Mapper) add(route *description) error {
	pos := len(m.services)
	m.services = append(m.services, route)
	for _, endpoint := range route.endpoints {
		old, updated := m.routers.Insert(route.prefix+endpoint, pos)
		if updated {
			return fmt.Errorf("%w, old: %#v , new: %s, endpoint: `%s`",
				ErrDuplicateEndpoint, old, route.name, endpoint,
			)
		}
	}
	return nil
}

func (m *Mapper) peekDefault() (*description, bool) {
	return m.def, m.def != nil
}

func (m *Mapper) peekServices() []*description {
	if len(m.services) == 0 {
		return []*description{}
	}
	out := make([]*description, len(m.services))
	copy(out, m.services)
	return out
}

func NewMapperFromConfig(cfg Config) (*Mapper, error) {
	srv := &Mapper{
		services: make([]*description, 0, len(cfg.Services)),
		routers:  radix.New(),
	}
	for _, svc := range cfg.Services {
		if len(svc.Endpoints) == 0 {
			return nil, fmt.Errorf("service map service %s %w", svc.Name, ErrNoEndpoints)
		}
		address, err := url.Parse(svc.URL)
		if err != nil {
			return nil, fmt.Errorf("parse %s address %s %w", svc.Name, svc.URL, err)
		}
		desc := &description{
			name:      svc.Name,
			address:   svc.URL,
			auth:      svc.Auth,
			prefix:    svc.Prefix,
			endpoints: svc.Endpoints,
			versions:  svc.Versions,
		}
		switch address.Scheme {
		case "http", "https":
			desc.transport = REST
		case "grpc":
			desc.transport = GRPC
		default:
			return nil, fmt.Errorf("svc %s url %s: %w", svc.Name, svc.URL, ErrUnknownTransport)
		}
		if err = srv.add(desc); err != nil {
			return nil, fmt.Errorf("failed to add service `%s` to routes: %w", svc.Name, err)
		}
	}
	if cfg.Default.URL != "" {
		srv.def = &description{
			name:      cfg.Default.Name,
			address:   cfg.Default.URL,
			auth:      cfg.Default.Auth,
			prefix:    cfg.Default.Prefix,
			transport: REST,
			versions:  cfg.Default.Versions,
			endpoints: cfg.Default.Endpoints,
		}
		if len(cfg.Default.Endpoints) != 0 {
			err := srv.add(srv.def)
			if err != nil {
				return nil, fmt.Errorf("failed to add default service `%s` to routes: %w", cfg.Default.Name, err)
			}
		} else {
			srv.services = append(srv.services, srv.def)
		}
	}

	return srv, nil
}

type Config struct {
	Services []configServiceDescription `yaml:"services"`
	Default  configServiceDescription   `yaml:"default"`
}

type configServiceDescription struct {
	Name      string           `yaml:"service"`
	URL       string           `yaml:"url"`
	Auth      bool             `yaml:"auth"`
	Prefix    string           `yaml:"prefix"`
	Endpoints []string         `yaml:"endpoints"`
	Versions  []ServiceVersion `yaml:"versions"`
}

func NewMapperFromYamlFile(path string) (*Mapper, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open services map file %s: %w", path, err)
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return NewMapperFromConfig(cfg)
}

func NewMapper() *Mapper {
	return &Mapper{
		services: []*description{},
		routers:  radix.New(),
	}
}
