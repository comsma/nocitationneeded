package config

type HCaptcha struct {
	SiteKey string `koanf:"sitekey"`
	Secret  string `koanf:"secret"`
}
