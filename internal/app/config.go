package app

type Config struct {
	TgToken   string `envconfig:"TG_TOKEN"`
	Once      bool   `envconfig:"ONCE"`
	DryRun    bool   `envconfig:"DRY_RUN"`
	SkipToday bool   `envconfig:"SKIP_TODAY"` // useful when doing multiple deployments per day
}
