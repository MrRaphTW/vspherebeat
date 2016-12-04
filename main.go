package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"git.teamwork.net/BeatsTeamwork/vspherebeat/beater"
)

func main() {
	err := beat.Run("vspherebeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
