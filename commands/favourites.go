package commands

func SaveFavourite(tool, name string, args []string) error {
	cfg, err := newToolsConfig()
	if err != nil {
		return err
	}
	return cfg.SaveFavourite(tool, name, args)
}
