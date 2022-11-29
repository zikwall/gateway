package gateway

type ServiceVersion struct {
	Version string `yaml:"version"`
	URL     string `yaml:"url"`
}

type Description interface {
	Transport() Transport
	Name() string
	Address() string
	Auth() bool
	Prefix() string
	Endpoints() []string
	Versions() []ServiceVersion
}

type description struct {
	transport Transport
	name      string
	address   string
	auth      bool
	prefix    string
	endpoints []string
	versions  []ServiceVersion
}

func (s *description) Transport() Transport {
	return s.transport
}

func (s *description) Name() string {
	return s.name
}

func (s *description) Address() string {
	return s.address
}

func (s *description) Auth() bool {
	return s.auth
}

func (s *description) Prefix() string {
	return s.prefix
}

func (s *description) Endpoints() []string {
	return s.endpoints
}

func (s *description) Versions() []ServiceVersion {
	return s.versions
}
