package options

import (
	"github.com/spf13/pflag"
)

var _ IOptions = (*NacosOptions)(nil)

type NacosOptions struct {
	Addr     string `json:"addr,omitempty" mapstructure:"addr"`
	Scheme   string `json:"scheme,omitempty" mapstructure:"scheme"`
	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`
}

func NewNacosOptions() *NacosOptions {
	return &NacosOptions{
		Addr:     "127.0.0.1:8848",
		Scheme:   "http",
		Username: "",
		Password: "",
	}
}

func (o *NacosOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (o *NacosOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, "nacos.addr", o.Addr, ""+
		"Addr is the address of the nacos server.")

	fs.StringVar(&o.Scheme, "nacos.scheme", o.Scheme, ""+
		"Scheme is the URI scheme for the nacos server.")

	fs.StringVar(&o.Username, "nacos.username", o.Username, ""+
		"Username for access to nacos server.")

	fs.StringVar(&o.Password, "nacos.password", o.Password, ""+
		"Password for access to nacos server.")
}
