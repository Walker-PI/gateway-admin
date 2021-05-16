package conf

type DBOptional struct {
	DriverName   string
	User         string
	Password     string
	DBHostname   string
	DBPort       string
	DBName       string
	DBCharset    string
	Timeout      string
	ReadTimeout  string
	WriteTimeout string
}

func GetDefaultDBOptional() *DBOptional {
	return &DBOptional{
		DriverName:   DBConf.DriverName,
		User:         DBConf.User,
		Password:     DBConf.Password,
		DBHostname:   DBConf.DBHostname,
		DBPort:       DBConf.DBPort,
		DBName:       DBConf.DBName,
		DBCharset:    "utf8",
		Timeout:      "1000ms",
		ReadTimeout:  "2.0s",
		WriteTimeout: "5.0s",
	}
}
