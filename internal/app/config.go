package app

type Config struct {
	TgToken string `envconfig:"TG_TOKEN"`
	DryRun  bool   `envconfig:"DRY_RUN"`
}
