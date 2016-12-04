package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	//  "github.com/vmware/govmomi/units"
	"context"
	"flag"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	//"text/tabwriter"
)

type cluster struct {
	dc          string
	name        string
	totalCPU    int16
	totalMemory int64
	nbHosts     int32
	path        string
	vsphereType string //Put "Cluster" here
}

func (theCluster cluster) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"totalCPU\":\"%d\", \"totalMemory\":\"%d\", \"nbHosts\":\"%d\", \"path\":\"%s\", \"vsphereType\":\"%s\"}\n", theCluster.name, theCluster.dc, theCluster.totalCPU, theCluster.totalMemory, theCluster.nbHosts, theCluster.path, theCluster.vsphereType)
}

type vm struct {
	name        string
	dc          string
	path        string
	cluster     string
	cpuLimit    int32
	diskLimit   int64
	memoryLimit int32
	vsphereType string //Put "VirtualMachine" here
}

func (theVM vm) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"path\":\"%s\", \"cluster\":\"%s\", \"cpuLimit\":\"%d\", \"memoryLimit\":\"%d\", \"diskLimit\":\"%d\", \"vsphereType\":\"%s\"}\n", theVM.name, theVM.dc, theVM.path, theVM.cluster, theVM.cpuLimit, theVM.memoryLimit, theVM.diskLimit, theVM.vsphereType)
}

type datastore struct {
	dc          string
	name        string
	capacity    int64
	freeSpace   int64
	path        string
	vsphereType string //Put "DataStore" here
}

func (theDS datastore) jsonRenderOnScreen() {
	fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"capacity\":\"%d\", \"freespace\":\"%d\", \"path\":\"%s\", \"vsphereType\":\"%s\"}\n", theDS.name, theDS.dc, theDS.capacity, theDS.freeSpace, theDS.path, theDS.vsphereType)
}

const (
	//URL : URL of the VSphere
	URL = "https://vstack-028.cloud-temple.com/sdk"
	//UserName  : Username to use for this vsphere
	UserName = "Florian.thoni@vsphere.local"
	// Password : Password to use for this vsphere
	Password = "tKOnvAb1KVcl<3"
	// Insecure : security to put in this vsphere connection
	Insecure = false
)

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
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
			storageinfos := vmdet.Storage.PerDatastoreUsage
			var alldiskspace int64
			alldiskspace = 0
			for _, storageinfo := range storageinfos {
				alldiskspace += storageinfo.Committed + storageinfo.Uncommitted
			}

			myVM := vm{name: vmoname, dc: dc.Name(), path: path, cluster: ownerName, cpuLimit: cpu, memoryLimit: memory, diskLimit: alldiskspace, vsphereType: "VirtualMachine"}
			return myVM
		}
		//TODO: Gérer ça comme une erreur
		fmt.Printf("This case has not been handled : owner is not a ClusterComputeResource!\n")
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
			//TODO: Gérer ça comme une erreur
			fmt.Printf("\nType for this child is not managed : %s\n", folderchild)
		}
	}
	return res
}

func getAllClusterInfo(ctx context.Context, c *govmomi.Client, f *find.Finder, dc *object.Datacenter) []cluster {
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
		myCluster := cluster{dc: dc.Name(), name: ccr.Name(), totalCPU: ressources.NumCpuThreads, totalMemory: ressources.TotalMemory, nbHosts: ressources.NumEffectiveHosts, path: path, vsphereType: "Cluster"}
		clusters = append(clusters, myCluster)
	}
	return clusters
}

func getAllDSInfo(ctx context.Context, c *govmomi.Client, f *find.Finder, dc *object.Datacenter) []datastore {
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
		myDS := datastore{dc: dc.Name(), name: dst.Summary.Name, capacity: dst.Summary.Capacity, freeSpace: dst.Summary.FreeSpace, path: path, vsphereType: "DataStore"}
		datastores = append(datastores, myDS)
	}
	return datastores
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", URL)
	var urlFlag = flag.String("url", URL, urlDescription)
	u, err := url.Parse(*urlFlag)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	key := []byte("SaltKey")
	plaintext := []byte("totopopo")
	cipthertex
	u.User = url.UserPassword(UserName, Password)
	c, err := govmomi.NewClient(ctx, u, false)
	if err != nil {
		log.Fatal(err)
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
		clusters := getAllClusterInfo(ctx, c, f, dc)
		for _, cluster := range clusters {
			cluster.jsonRenderOnScreen()
		}

		datastores := getAllDSInfo(ctx, c, f, dc)
		for _, ds := range datastores {
			ds.jsonRenderOnScreen()
		}
		// Now we get the VM Info by a path exploration
		folders, err := dc.Folders(ctx)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		vmfolder := folders.VmFolder
		vms := explorevmfolder(ctx, c, vmfolder, path, dc)
		for _, vm := range vms {
			vm.jsonRenderOnScreen()
		}
	}
}
