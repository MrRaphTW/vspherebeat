//Package beater : no unit test on this part, this would require an available vsphere.
package beater

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
)

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
	dss, err := f.DatastoreList(ctx, "*")
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	pc := property.DefaultCollector(c.Client)
	var datastores []datastore
	for _, ds := range dss {
		refds := ds.Reference()
		var dst mo.Datastore
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
