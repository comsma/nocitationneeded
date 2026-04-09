package config

type HCaptcha struct {
	SiteKey string `yaml:"sitekey"`
	Secret  string `yaml:"secret"`
}
