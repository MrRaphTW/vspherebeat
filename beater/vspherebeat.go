//Package beater: no unit test on this part
package beater

import (
	"context"
	"flag"
	"fmt"
	"net/url"

	"git.teamwork.net/BeatsTeamwork/vspherebeat/config"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/robfig/cron"
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

func (bt *Vspherebeat) RunOnce(b *beat.Beat, ctx context.Context, u *url.URL) error {

	c, err := govmomi.NewClient(ctx, u, bt.config.Insecure)
	if err != nil {
		return err
	}
	// Here we start parsing the whole DC lists in this client
	f := find.NewFinder(c.Client, true)

	dcs, err := f.DatacenterList(ctx, "*")
	if err != nil {
		return err
	}

	var path string

	for _, dc := range dcs {
		fmt.Printf("We found this DC : [%s]\n", dc)
		f.SetDatacenter(dc)
		path = (*dc).Common.InventoryPath
		fmt.Printf("Path - [%s]\n", path)
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
			return err
		}
		vmfolder := folders.VmFolder
		vms := explorevmfolder(ctx, c, vmfolder, path, dc)
		for _, vm := range vms {
			event := vm.eventRender(b)
			bt.client.PublishEvent(event)
			logp.Info("Event sent")
		}

	}
	return nil
}

// Run is the main loop
func (bt *Vspherebeat) Run(b *beat.Beat) error {
	logp.Info("vspherebeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logp.Info("Start Running once")

	var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", bt.config.URL)
	var urlFlag = flag.String("url", bt.config.URL, urlDescription)
	u, err := url.Parse(*urlFlag)
	if err != nil {
		return err
	}
	var password string
	if bt.config.EncPassword {
		password, err = decryptString(bt.config.Password, encryptionKey)
		if err != nil {
			return err
		}
	} else {
		password = bt.config.Password
	}
	u.User = url.UserPassword(bt.config.UserName, password)

	cron := cron.New()

	cron.AddFunc(bt.config.Cron, func() { bt.RunOnce(b, ctx, u) })
	cron.Start()
	logp.Info("Started the CRON. ")

	for {
		select {
		case <-bt.done:
			cron.Stop()
			return nil
		}
	}
}

// Stop handles what is needed at stop moment
func (bt *Vspherebeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
