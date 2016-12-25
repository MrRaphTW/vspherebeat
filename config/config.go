// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

// Config structure that handles all options for the beat and especially one to connect to vsphere
type Config struct {
	Cron             string `config:"Cron"`
	URL              string `config:"URL"`
	UserName         string `config:"UserName"`
	Password         string `config:"Password"`
	EncPassword      bool   `config:"EncPassword"`
	Insecure         bool   `config:"Insecure"`
	PrecoCPUPercent  int    `config:"PrecoCPUPercent"`
	PrecoRAMPercent  int    `config:"PrecoRAMPercent"`
	PrecoDiskPercent int    `config:"PrecoDiskPercent"`
}

// DefaultConfig is the object to have the default configuration if a parameter is missing.
var DefaultConfig = Config{
	Cron:             "@daily",
	URL:              "",
	UserName:         "",
	Password:         "",
	EncPassword:      false,
	Insecure:         false,
	PrecoCPUPercent:  100,
	PrecoRAMPercent:  100,
	PrecoDiskPercent: 100,
}
