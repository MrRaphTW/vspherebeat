// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

// Config structure that handles all options for the beat and especially one to connect to vsphere
type Config struct {
	Period           time.Duration `config:"period"`
	URL              string        `config:"URL"`
	UserName         string        `config:"UserName"`
	Password         string        `config:"Password"`
	Insecure         bool          `config:"Insecure"`
	PrecoCPUPercent  int           `config:"PrecoCPUPercent"`
	PrecoRAMPercent  int           `config:"PrecoRAMPercent"`
	PrecoDiskPercent int           `config:"PrecoDiskPercent"`
}

// DefaultConfig is the object to have the default configuration if a parameter is missing.
var DefaultConfig = Config{
	Period:           24 * 7 * time.Hour,
	URL:              "",
	UserName:         "",
	Password:         "",
	Insecure:         false,
	PrecoCPUPercent:  100,
	PrecoRAMPercent:  100,
	PrecoDiskPercent: 100,
}
