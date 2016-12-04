package beater

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"golang.org/x/crypto/blowfish"

	"git.teamwork.net/BeatsTeamwork/vspherebeat/config"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
)

// Vspherebeat is the main Vspherebeat structure
type Vspherebeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

type cluster struct {
	dc                string
	name              string
	totalCPU          int16
	totalMemory       int64
	nbHosts           int32
	path              string
	cpuOverallocPreco int
	ramOverallocPreco int
}

func (theCluster cluster) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"totalCPU\":\"%d\", \"totalMemory\":\"%d\", \"nbHosts\":\"%d\", \"path\":\"%s\", \"vsphereType\":\"Cluster\"}\n", theCluster.name, theCluster.dc, theCluster.totalCPU, theCluster.totalMemory, theCluster.nbHosts, theCluster.path)
}

func (theCluster cluster) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":        common.Time(time.Now()),
		"type":              b.Name,
		"dc":                theCluster.dc,
		"name":              theCluster.name,
		"totalCPU":          theCluster.totalCPU,
		"totalMemory":       theCluster.totalMemory,
		"nbHosts":           theCluster.nbHosts,
		"path":              theCluster.path,
		"cpuOverallocPreco": theCluster.cpuOverallocPreco,
		"ramOverallocPreco": theCluster.ramOverallocPreco,
		"vsphereType":       "Cluster",
	}
	return event
}

type vm struct {
	name        string
	dc          string
	path        string
	cluster     string
	cpuLimit    int32
	memoryLimit int32
	diskLimit   int64
}

func (theVM vm) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"path\":\"%s\", \"cluster\":\"%s\", \"cpuLimit\":\"%d\", \"memoryLimit\":\"%d\", \"diskLimit\":\"%d\", \"vsphereType\":\"VirtualMachine\"}\n", theVM.name, theVM.dc, theVM.path, theVM.cluster, theVM.cpuLimit, theVM.memoryLimit, theVM.diskLimit)
}

func (theVM vm) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":  common.Time(time.Now()),
		"type":        b.Name,
		"dc":          theVM.dc,
		"path":        theVM.path,
		"cluster":     theVM.cluster,
		"cpuLimit":    theVM.cpuLimit,
		"memoryLimit": theVM.memoryLimit,
		"diskLimit":   theVM.diskLimit,
		"vsphereType": "VirtualMachine",
	}
	return event
}

type datastore struct {
	dc                 string
	name               string
	capacity           int64
	freeSpace          int64
	path               string
	diskOverallocPreco int
}

func (theDS datastore) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"capacity\":\"%d\", \"freespace\":\"%d\", \"path\":\"%s\", \"vsphereType\":\"DataStore\"}\n", theDS.name, theDS.dc, theDS.capacity, theDS.freeSpace, theDS.path)
}

func (theDS datastore) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":         common.Time(time.Now()),
		"type":               b.Name,
		"dc":                 theDS.dc,
		"name":               theDS.name,
		"capacity":           theDS.capacity,
		"freeSpace":          theDS.freeSpace,
		"path":               theDS.path,
		"diskOverallocPreco": theDS.diskOverallocPreco,
		"vsphereType":        "DataStore",
	}
	return event
}

func getvminfo(ctx context.Context, c *govmomi.Client, theVM *object.VirtualMachine, path string, dc *object.Datacenter) vm {
	vmoname, err := theVM.ObjectName(ctx)
	path += "/"
	path += vmoname
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	pc := property.DefaultCollector(c.Client)
	// Convert VM into list of references
	vmref := theVM.Reference()

	var vmdet mo.VirtualMachine

	err = pc.RetrieveOne(ctx, vmref, []string{"config", "summary", "storage"}, &vmdet)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	if !vmdet.Config.Template {
		//fmt.Printf("Path : [%s] - ObjectName : [%s]\n", path, vmoname)
		//fmt.Printf("Not template\n")
		vmRP, err := theVM.ResourcePool(ctx)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		vmRPref := vmRP.Reference()

		var vmRPdet mo.ResourcePool

		err = pc.RetrieveOne(ctx, vmRPref, []string{"owner"}, &vmRPdet)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		ownerRef := vmRPdet.Owner.Reference()
		if ownerRef.Type == "ClusterComputeResource" {
			owner := object.NewClusterComputeResource(c.Client, ownerRef)
			ownerName, err := owner.ObjectName(ctx)
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			//fmt.Printf("Cluster Name :%s\n", ownerName)
			cpu := vmdet.Summary.Config.NumCpu
			memory := vmdet.Summary.Config.MemorySizeMB
			var alldiskspace int64
			alldiskspace = 0
			storageinfos := vmdet.Storage.PerDatastoreUsage
			for _, storageinfo := range storageinfos {
				alldiskspace += storageinfo.Committed + storageinfo.Uncommitted
			}
			myVM := vm{name: vmoname, dc: dc.Name(), path: path, cluster: ownerName, cpuLimit: cpu, memoryLimit: memory, diskLimit: alldiskspace}
			return myVM
		}
		errmess := fmt.Sprintf("This case has not been handled : owner is not a ClusterComputeResource!\n")
		log.Fatal(errors.New(errmess))
	}
	return vm{}
}

func explorevmfolder(ctx context.Context, c *govmomi.Client, folder *object.Folder, path string, dc *object.Datacenter) []vm {
	currentfoldername, err := folder.ObjectName(ctx)
	path += "/"
	path += currentfoldername
	//fmt.Printf("Currently exploring folder [%s] at location (%s)\n", currentfoldername, path)
	folderchildren, err := folder.Children(ctx)
	var res []vm
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	if len(folderchildren) == 0 {
		//fmt.Printf("This folder is empty !\n")
	}
	for _, folderchild := range folderchildren {
		switch folderchild.(type) {
		case *object.VirtualMachine:
			//fmt.Printf("\nType of the found child here is : object.VirtualMachine\n")
			newvm := getvminfo(ctx, c, folderchild.(*object.VirtualMachine), path, dc)
			emptyvm := vm{}
			if newvm != emptyvm {
				res = append(res, newvm)
			}
		case *object.Folder:
			//fmt.Printf("\nType of the found child here is : object.Folder\n")
			newvms := explorevmfolder(ctx, c, folderchild.(*object.Folder), path, dc)
			res = append(res, newvms...)
		default:
			errmess := fmt.Sprintf("Type for this child is not managed : %s", folderchild)
			log.Fatal(errors.New(errmess))
		}
	}
	return res
}

func getAllClusterInfo(ctx context.Context, c *govmomi.Client, f *find.Finder, dc *object.Datacenter, cpuOverallocPreco int, ramOverallocPreco int) []cluster {
	ccrs, err := f.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	pc := property.DefaultCollector(c.Client)
	var clusters []cluster
	for _, ccr := range ccrs {
		refccr := ccr.Reference()
		var ccrt mo.ClusterComputeResource
		err = pc.RetrieveOne(ctx, refccr, []string{"summary"}, &ccrt)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		ressources := ccrt.Summary.GetComputeResourceSummary()
		path := ccr.ComputeResource.Common.InventoryPath
		myCluster := cluster{dc: dc.Name(), name: ccr.Name(), totalCPU: ressources.NumCpuThreads, totalMemory: ressources.TotalMemory, nbHosts: ressources.NumEffectiveHosts, path: path, cpuOverallocPreco: cpuOverallocPreco, ramOverallocPreco: ramOverallocPreco}
		clusters = append(clusters, myCluster)
	}
	return clusters
}

func getAllDSInfo(ctx context.Context, c *govmomi.Client, f *find.Finder, dc *object.Datacenter, diskOverallocPreco int) []datastore {
	dss, err := f.DatastoreClusterList(ctx, "*")
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	pc := property.DefaultCollector(c.Client)
	var datastores []datastore
	for _, ds := range dss {
		refds := ds.Reference()
		var dst mo.StoragePod
		err = pc.RetrieveOne(ctx, refds, []string{"summary"}, &dst)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		path := ds.Common.InventoryPath
		myDS := datastore{dc: dc.Name(), name: dst.Summary.Name, capacity: dst.Summary.Capacity, freeSpace: dst.Summary.FreeSpace, path: path, diskOverallocPreco: diskOverallocPreco}
		datastores = append(datastores, myDS)
	}
	return datastores
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
	myCipher, err := blowfish.NewCipher("tutu")
	u.User = url.UserPassword(bt.config.UserName, bt.config.Password)
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
