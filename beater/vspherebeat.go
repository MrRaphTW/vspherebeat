//Package beater: no unit test on this part
package beater

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"git.teamwork.net/BeatsTeamwork/vspherebeat/config"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

var encryptionKey string

// Vspherebeat is the main Vspherebeat structure
type Vspherebeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// New Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Vspherebeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

// Run is the main loop
func (bt *Vspherebeat) Run(b *beat.Beat) error {
	logp.Info("vspherebeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", bt.config.URL)
	var urlFlag = flag.String("url", bt.config.URL, urlDescription)
	u, err := url.Parse(*urlFlag)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	var password string
	if bt.config.EncPassword {
		password, err = decryptString(bt.config.Password, encryptionKey)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		password = bt.config.Password
	}
	u.User = url.UserPassword(bt.config.UserName, password)
	c, err := govmomi.NewClient(ctx, u, false)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		// Here we start parsing the whole DC lists in this client
		f := find.NewFinder(c.Client, true)

		dcs, err := f.DatacenterList(ctx, "*")
		if err != nil {
			log.Fatal(err)
		}

		var path string

		for _, dc := range dcs {
			//fmt.Printf("We found this DC : [%s]\n", dc)
			f.SetDatacenter(dc)
			path = (*dc).Common.InventoryPath
			//fmt.Printf("Path - [%s]\n", path)
			// First we get the whole Clusters information in this DC
			clusters := getAllClusterInfo(ctx, c, f, dc, bt.config.PrecoCPUPercent, bt.config.PrecoRAMPercent)
			for _, cluster := range clusters {
				event := cluster.eventRender(b)
				bt.client.PublishEvent(event)
				logp.Info("Event sent")
			}

			datastores := getAllDSInfo(ctx, c, f, dc, bt.config.PrecoDiskPercent)
			for _, ds := range datastores {
				event := ds.eventRender(b)
				bt.client.PublishEvent(event)
				logp.Info("Event sent")
			}
			// Now we get the VM Info by a path exploration
			folders, err := dc.Folders(ctx)
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			vmfolder := folders.VmFolder
			vms := explorevmfolder(ctx, c, vmfolder, path, dc)
			for _, vm := range vms {
				event := vm.eventRender(b)
				bt.client.PublishEvent(event)
				logp.Info("Event sent")
			}
		}

	}
}

// Stop handles what is needed at stop moment
func (bt *Vspherebeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
