package config

type PostgresConfig struct {
	ConectionFormat string `yaml:"db_con_format"`
	Host            string `yaml:"db_host"`
	Port            string `yaml:"db_port"`
	User            string `yaml:"db_user"`
	Password        string `yaml:"db_pass"`
	Database        string `yaml:"db_name"`
	UserTable       string `yaml:"db_tbl_user"`
	CatogoryTable   string `yaml:"db_tbl_category"`
}