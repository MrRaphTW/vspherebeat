package main


import (
  "fmt"
  "net/url"
  "log"
  "os"
  "github.com/vmware/govmomi"
  "github.com/vmware/govmomi/find"
  "github.com/vmware/govmomi/object"
//  "github.com/vmware/govmomi/units"
  "github.com/vmware/govmomi/property"
  "github.com/vmware/govmomi/vim25/mo"
  "github.com/vmware/govmomi/vim25/types"
  "context"
  "flag"
  "reflect"
  //"text/tabwriter"
)
type cluster struct {
    dc string
    name string
    totalCPU int64
    totalMemory int64
}
func (theCluster cluster) jsonRenderOnScreen() {
    fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"totalCPU\":\"%d\", \"totalMemory\":\"%d\"}\n", theCluster.name, theCluster.dc, theCluster.totalCPU, theCluster.totalMemory)
}

type vm struct {
    name string
    dc string
    path string
}

func (theVM vm) jsonRenderOnScreen() {
    fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"path\":\"%s\"}\n", theVM.name, theVM.dc, theVM.path)
}

type datastore struct {
    dc string
    name string
    capacity int64
    freeSpace int64
}

func (theDS datastore) jsonRenderOnScreen() {
    fmt.Printf("{\"name\":\"%s\", \"dc\":\"%s\", \"capacity\":\"%d\", \"freespace\":\"%d\"}\n", theDS.name, theDS.dc, theDS.capacity, theDS.freeSpace)
}

const (
    URL = "https://vstack-028.cloud-temple.com/sdk"
    UserName = "Florian.thoni@vsphere.local"
    Password = "tKOnvAb1KVcl<3"
    Insecure = false
)

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}



func getvminfo(ctx context.Context, c *govmomi.Client, vm *object.VirtualMachine,path string){
    vmoname, err := vm.ObjectName(ctx)
    path += "/"
    path += vmoname
    if err != nil {
        fmt.Printf("%s\n", err)
    }

    fmt.Printf("Path : [%s] - ObjectName : [%s]", path, vmoname)
/*
    pc := property.DefaultCollector(c.Client)
    // Convert VM into list of references
    vmref := vm.Reference()
    refs := []types.ManagedObjectReference{vmref}

    var vmt []mo.VirtualMachine


    err = pc.Retrieve(ctx, refs, []string{"summary"}, &vmt)
    if err != nil {
        log.Fatal(err)
    }

    for _, vmdet := range vmt {
        test := reflect.ValueOf(vmdet)//.Elem()
        fmt.Printf("\n\n%s\n", test)
        fmt.Printf("    %s\n", test.Type())

        if true {
            fmt.Printf("\nProperties\n")
            for i:= 0; i < test.NumField();i++ {
                fmt.Printf("    %s - %s\n", test.Type().Field(i).Name , test.Type().Field(i).Type)
            }
        }
        fmt.Printf("\nMethods\n")
        for i:= 0; i < test.NumMethod();i++ {
            fmt.Printf("    %s\n", test.Type().Method(i).Name )
        }


    }
*/
/*    props, _ := vm.(*object.VirtualMachine).Properties(ctx)
    test := reflect.ValueOf(props).Elem
    fmt.Printf("\n\n%s\n", test)
    fmt.Printf("    %s\n", test.Type())

    if true {
        fmt.Printf("\nProperties\n")
        for i:= 0; i < test.NumField();i++ {
            fmt.Printf("    %s - %s\n", test.Type().Field(i).Name , test.Type().Field(i).Type)
        }
    }
    fmt.Printf("\nMethods\n")
    for i:= 0; i < test.NumMethod();i++ {
        fmt.Printf("    %s\n", test.Type().Method(i).Name )
    }
    */
}

func explorefolder(ctx context.Context, c *govmomi.Client, folder *object.Folder, path string) {
    currentfoldername, err := folder.ObjectName(ctx)
    path += "/"
    path += currentfoldername
    fmt.Printf("Currently exploring folder [%s] at location (%s)", currentfoldername, path)
    folderchildren, err := folder.Children(ctx)
    if err != nil {
        fmt.Printf("%s\n", err)
    }
    if len(folderchildren) == 0 {
        fmt.Printf ("This folder is empty !\n")
    }
    for _, folderchild := range folderchildren {
        switch folderchild.(type) {
            case *object.VirtualMachine:
                fmt.Printf("\nType of the found child here is : object.VirtualMachine\n")
                getvminfo(ctx, c, folderchild.(*object.VirtualMachine), path)



            case *object.Folder:
                fmt.Printf("\nType of the found child here is : object.Folder\n")
                explorefolder(ctx, c, folderchild.(*object.Folder), path)
            default:
                fmt.Printf("\nType for this child is not managed : %s\n", folderchild)
        }
    }
}

func getAllClusterInfo(ctx context.Context, c *govmomi.Client, f *find.Finder) {
    ccrs, err := f.ClusterComputeResourceList(ctx, "*")
    if err != nil {
        fmt.Printf("%s\n", err)
    }

    pc := property.DefaultCollector(c.Client)

    for _, ccr := range ccrs {
        refccr := ccr.Reference()
        var ccrt mo.ClusterComputeResource
        err = pc.RetrieveOne(ctx, refccr, []string{"summary"}, &ccrt)
        if err != nil {
            fmt.Printf("%s\n", err)
        }
        fmt.Printf("Name: %s\n",ccr.Name())

    }



    if true {
        for _, ccr := range ccrt {
            //var summary types.ClusterComputeResourceSummary
            ressources := ccr.Summary.GetComputeResourceSummary()
            used := ccr.Summary.(*types.ClusterComputeResourceSummary).UsageSummary
            //usage := ressources.UsageSummary
            fmt.Printf("    Ressources")
            fmt.Printf("        TotalCpu : %d\n", ressources.TotalCpu )
            fmt.Printf("        TotalMemory : %d\n", ressources.TotalMemory )
            fmt.Printf("        NumCpuCores : %d\n", ressources.NumCpuCores )
            fmt.Printf("        NumCpuThreads : %d\n", ressources.NumCpuThreads)
            fmt.Printf("        EffectiveCpu : %d\n", ressources.EffectiveCpu )
            fmt.Printf("        EffectiveMemory : %d\n", ressources.EffectiveMemory )
            fmt.Printf("        NumHosts : %d\n", ressources.NumHosts )
            fmt.Printf("        NumEffectiveHosts : %d\n", ressources.NumEffectiveHosts )
            fmt.Printf("        OverallStatus : %s\n", ressources.OverallStatus )

            fmt.Printf("        TotalCpuCapacityMhz : %d\n", used.TotalCpuCapacityMhz)
            fmt.Printf("        TotalMemCapacityMB : %d\n", used.TotalMemCapacityMB)
            fmt.Printf("        CpuReservationMhz : %d\n", used.CpuReservationMhz)
            fmt.Printf("        MemReservationMB : %d\n", used.MemReservationMB)
            fmt.Printf("        PoweredOffCpuReservationMhz : %d\n", used.PoweredOffCpuReservationMhz)
            fmt.Printf("        PoweredOffMemReservationMB : %d\n", used.PoweredOffMemReservationMB)
            fmt.Printf("        CpuDemandMhz : %d\n", used.CpuDemandMhz)
            fmt.Printf("        MemDemandMB : %d\n", used.MemDemandMB)
            fmt.Printf("        StatsGenNumber : %d\n", used.StatsGenNumber)
            fmt.Printf("        CpuEntitledMhz : %d\n", used.CpuEntitledMhz)
            fmt.Printf("        MemEntitledMB : %d\n", used.MemEntitledMB)
            fmt.Printf("        PoweredOffVmCount : %d\n", used.PoweredOffVmCount)
            fmt.Printf("        TotalVmCount : %d\n", used.TotalVmCount)

            test := reflect.ValueOf(ccr)//.Elem()
            fmt.Printf("\n\n%s\n", test)
            fmt.Printf("    %s\n", test.Type())

            if true {
                fmt.Printf("\nProperties\n")
                for i:= 0; i < test.NumField();i++ {
                    fmt.Printf("    %s - %s\n", test.Type().Field(i).Name , test.Type().Field(i).Type)
                }
            }
            fmt.Printf("\nMethods\n")
            for i:= 0; i < test.NumMethod();i++ {
                fmt.Printf("    %s\n", test.Type().Method(i).Name )
            }
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", URL)
    var urlFlag = flag.String("url", URL, urlDescription)
    u, err := url.Parse(*urlFlag)
    if err != nil {
        fmt.Println("%s", err)
    }
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

    var path string;

    for _, dc := range dcs {
        //fmt.Printf("We found this DC : [%s]\n", dc)
        f.SetDatacenter(dc)
        path = (*dc).Common.InventoryPath
        fmt.Printf("Path - [%s]\n", path)
        getAllClusterInfo(ctx, c, f)
    }
}



func old() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()


    fmt.Printf("Hello World\n")     // or resp.String() or string(resp.Body())
    var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", URL)
    var urlFlag = flag.String("url", URL, urlDescription)
    u, err := url.Parse(*urlFlag)
    if err != nil {
        log.Fatal(err)
    }
    u.User = url.UserPassword(UserName, Password)
    c, err := govmomi.NewClient(ctx, u, false)
    if err != nil {
        log.Fatal(err)
    }
    f := find.NewFinder(c.Client, true)


    dcs, err := f.DatacenterList(ctx, "*")
    if err != nil {
        log.Fatal(err)
    }
    var path string;

    for _, dc := range dcs {
        fmt.Printf("We found this DC : [%s]\n", dc)
        f.SetDatacenter(dc)
        path = (*dc).Common.InventoryPath
        fmt.Printf("The path for it is : [%s]\n", path)

        if true {
            folders, err := dc.Folders(ctx)
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            vmfolder := folders.VmFolder
            explorefolder(ctx, c, vmfolder, path)
            /*
            vmfoldchildren, err := vmfolder.Children(ctx)
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            for _, vmfoldchild := range vmfoldchildren {
                vmfoldchildchildren, err := vmfoldchild.(*object.Folder).Children(ctx)
                if err != nil {
                    fmt.Printf("%s\n", err)
                }
                for _, vmfoldchildchild := range vmfoldchildchildren {
                    switch vmfoldchildchild.(type) {
                        case *object.VirtualMachine:
                            fmt.Printf("\nType here is : object.VirtualMachine\n")
                        case *object.Folder:
                            fmt.Printf("\nType here is : object.Folder\n")
                        default:
                            fmt.Printf("\nType here is not managed for %s\n", vmfoldchildchild)
                    }
                    test := reflect.ValueOf(vmfoldchildchild).Elem()
                    fmt.Printf("\n\n%s\n", test)
                    fmt.Printf("    %s\n", test.Type())

                    if true {
                        fmt.Printf("\nProperties\n")
                        for i:= 0; i < test.NumField();i++ {
                            fmt.Printf("    %s - %s\n", test.Type().Field(i).Name , test.Type().Field(i).Type)
                        }
                    }
                    fmt.Printf("\nMethods\n")
                    for i:= 0; i < test.NumMethod();i++ {
                        fmt.Printf("    %s\n", test.Type().Method(i).Name )
                    }
                }
            }
            */
        }

        if false {
            //fmt.Printf("\n\n%s\n\n", dc.Folders(ctx))
            /*
            //Je ne sais fichtre rien de ce que je peux faire avec ces folders !
            folders, err := dc.Folders(ctx)
            if err != nil {
                log.Fatal(err)
            }
            vmfold := folders.VmFolder
            */


            //ça ça nous donne donc les vm à la racine d'une dossier donné.
            fmt.Printf("Let get the VMs\n")
            vms, err := f.VirtualMachineList(ctx, "*")
            if err != nil {
                fmt.Printf("%s\n", err)
                //log.Fatal(err)
            }

            for _, vm := range vms {
                fmt.Printf("VM Here is the inventory Path : [%s]\n", vm.Common.InventoryPath)
                fmt.Printf("VM Here is the name : [%s]\n", vm.Common.Name())
            }

            rps, err := f.ResourcePoolList(ctx, "*")
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            for _, rp := range rps {
                fmt.Printf("%s\n", rp)
                fmt.Printf("RP Here is the inventory Path : [%s]\n", rp.Common.InventoryPath)
                fmt.Printf("RP Here is the name : [%s]\n", rp.Common.Name())
            }

            vps, err := f.VirtualAppList(ctx, "*")
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            for _, vp := range vps {
                fmt.Printf("%s\n", vp)
                fmt.Printf("VP Here is the inventory Path : [%s]\n", vp.Common.InventoryPath)
                fmt.Printf("VP Here is the name : [%s]\n", vp.Common.Name())
            }

            fls, err := f.FolderList(ctx, "*")
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            for _, fl := range fls {
                fmt.Printf("%s\n", fl)
                fmt.Printf("FL Here is the inventory Path : [%s]\n", fl.Common.InventoryPath)
                fmt.Printf("FL Here is the name : [%s]\n", fl.Common.Name())

            }

            ccrs, err := f.ClusterComputeResourceList(ctx, "*")
            if err != nil {
                fmt.Printf("%s\n", err)
            }
            for _, ccr := range ccrs {
                fmt.Printf("%s\n", ccr)
                fmt.Printf("CCR Here is the inventory Path : [%s]\n", ccr.Common.InventoryPath)
                fmt.Printf("CCR Here is the name : [%s]\n", ccr.Common.Name())


                //fmt.Printf("\n\n%s\n", .Common.InventoryPath)
                if false {
                    test := reflect.ValueOf(ccr).Elem()
                    fmt.Printf("\n\n%s\n", test)
                    fmt.Printf("\n\n%s\n", test.Type())

                    fmt.Printf("\nProperties\n")
                    for i:= 0; i < test.NumField();i++ {
                        fmt.Printf("%s - %s\n", test.Type().Field(i).Name , test.Type().Field(i).Type)
                    }
                    fmt.Printf("\nMethods\n")
                    for i:= 0; i < test.NumMethod();i++ {
                        fmt.Printf("%s\n", test.Type().Method(i).Name )
                    }
                }

            }
        }


    }

    vms2, err := f.VirtualMachineList(ctx,"/DC-ITX7/vm/DC-ITX7/%2f/tsm1-vee-itx7")
    if err != nil {
        fmt.Printf("%s\n", err)
        //log.Fatal(err)
    }

    for _, vm2 := range vms2 {
        fmt.Printf("%s\n", vm2)
    }

    fmt.Printf("\n\n####TESTS\n\n")

    //HERE CLUSTERCOMPUTERESOURCE - Not really useful since will only consider the max allocated for pool if pool exists.
            ccrs, err := f.ClusterComputeResourceList(ctx, "*")
            if err != nil {
                log.Fatal(err)
            }

            pc := property.DefaultCollector(c.Client)
            var refs3 []types.ManagedObjectReference

            for _, ccr := range ccrs {
                refs3 = append(refs3, ccr.Reference())
            }

            var ccrt []mo.ClusterComputeResource

            err = pc.Retrieve(ctx, refs3, []string{"summary"}, &ccrt)
            if err != nil {
                log.Fatal(err)
            }

            for _, ccr := range ccrt {
                //var summary types.ClusterComputeResourceSummary
                ressources := ccr.Summary.GetComputeResourceSummary()
                used := ccr.Summary.(*types.ClusterComputeResourceSummary).UsageSummary
                //usage := ressources.UsageSummary
                fmt.Printf("    Ressources")
                fmt.Printf("        TotalCpu : %d\n", ressources.TotalCpu )
                fmt.Printf("        TotalMemory : %d\n", ressources.TotalMemory )
                fmt.Printf("        NumCpuCores : %d\n", ressources.NumCpuCores )
                fmt.Printf("        NumCpuThreads : %d\n", ressources.NumCpuThreads)
                fmt.Printf("        EffectiveCpu : %d\n", ressources.EffectiveCpu )
                fmt.Printf("        EffectiveMemory : %d\n", ressources.EffectiveMemory )
                fmt.Printf("        NumHosts : %d\n", ressources.NumHosts )
                fmt.Printf("        NumEffectiveHosts : %d\n", ressources.NumEffectiveHosts )
                fmt.Printf("        OverallStatus : %s\n", ressources.OverallStatus )

                fmt.Printf("        TotalCpuCapacityMhz : %d\n", used.TotalCpuCapacityMhz)
                fmt.Printf("        TotalMemCapacityMB : %d\n", used.TotalMemCapacityMB)
                fmt.Printf("        CpuReservationMhz : %d\n", used.CpuReservationMhz)
                fmt.Printf("        MemReservationMB : %d\n", used.MemReservationMB)
                fmt.Printf("        PoweredOffCpuReservationMhz : %d\n", used.PoweredOffCpuReservationMhz)
                fmt.Printf("        PoweredOffMemReservationMB : %d\n", used.PoweredOffMemReservationMB)
                fmt.Printf("        CpuDemandMhz : %d\n", used.CpuDemandMhz)
                fmt.Printf("        MemDemandMB : %d\n", used.MemDemandMB)
                fmt.Printf("        StatsGenNumber : %d\n", used.StatsGenNumber)
                fmt.Printf("        CpuEntitledMhz : %d\n", used.CpuEntitledMhz)
                fmt.Printf("        MemEntitledMB : %d\n", used.MemEntitledMB)
                fmt.Printf("        PoweredOffVmCount : %d\n", used.PoweredOffVmCount)
                fmt.Printf("        TotalVmCount : %d\n", used.TotalVmCount)



            }




}



/*
        dss, err := f.DatastoreClusterList(ctx, "*")
        if err != nil {
            log.Fatal(err)
        }


        pc := property.DefaultCollector(c.Client)

        // Convert datastores into list of references
    	var refs []types.ManagedObjectReference
    	for _, ds := range dss {
    		refs = append(refs, ds.Reference())
    	}

        var dst []mo.StoragePod


        err = pc.Retrieve(ctx, refs, []string{"summary"}, &dst)
        if err != nil {
            log.Fatal(err)
        }


        for _, ds := range dst {
            fmt.Printf("    Storage Name : %s\n", ds.Summary.Name)
            //fmt.Printf("        Type : %s\n", ds.Summary.Type)
            fmt.Printf("        Capacity : %s\n", units.ByteSize(ds.Summary.Capacity))
            fmt.Printf("        FreeSpace : %s\n", units.ByteSize(ds.Summary.FreeSpace))
        }

/*
        crs, err := f.ComputeResourceList(ctx, "*")
        if err != nil {
            log.Fatal(err)
        }

        pc = property.DefaultCollector(c.Client)
        var refs2 []types.ManagedObjectReference

        for _, cr := range crs {
            refs2 = append(refs2, cr.Reference())
        }

        var crt []mo.ComputeResource

        err = pc.Retrieve(ctx, refs2, []string{"summary"}, &crt)
        if err != nil {
            log.Fatal(err)
        }

        for _, cr := range crt {
            ressources := cr.Summary.GetComputeResourceSummary()
            fmt.Printf("    Ressources")
            fmt.Printf("        TotalCpu : %d\n", ressources.TotalCpu )
            fmt.Printf("        TotalMemory : %d\n", ressources.TotalMemory )
            fmt.Printf("        NumCpuCores : %d\n", ressources.NumCpuCores )
            fmt.Printf("        NumCpuThreads : %d\n", ressources.NumCpuThreads)
            fmt.Printf("        EffectiveCpu : %d\n", ressources.EffectiveCpu )
            fmt.Printf("        EffectiveMemory : %d\n", ressources.EffectiveMemory )
            fmt.Printf("        NumHosts : %d\n", ressources.NumHosts )
            fmt.Printf("        NumEffectiveHosts : %d\n", ressources.NumEffectiveHosts )
            fmt.Printf("        OverallStatus : %s\n", ressources.OverallStatus )
        }*/
/*


    //fmt.Printf("%s", res)
    */

/*
    // Find one and only datacenter
    	dc, err := f.DefaultDatacenter(ctx)
    	if err != nil {
    		log.Fatal(err)
    	}

    	// Make future calls local to this datacenter
    	f.SetDatacenter(dc)

    	// Find datastores in datacenter
    	dss, err := f.DatastoreList(ctx, "*")
    	if err != nil {
    		log.Fatal(err)
    	}

    	pc := property.DefaultCollector(c.Client)

    	// Convert datastores into list of references
    	var refs []types.ManagedObjectReference
    	for _, ds := range dss {
    		refs = append(refs, ds.Reference())
    	}

    	// Retrieve summary property for all datastores
    	var dst []mo.Datastore
    	err = pc.Retrieve(ctx, refs, []string{"summary"}, &dst)
    	if err != nil {
    		log.Fatal(err)
    	}

    	// Print summary per datastore
    	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
    	fmt.Fprintf(tw, "Name:\tType:\tCapacity:\tFree:\n")
    	for _, ds := range dst {
    		fmt.Fprintf(tw, "%s\t", ds.Summary.Name)
    		fmt.Fprintf(tw, "%s\t", ds.Summary.Type)
    		fmt.Fprintf(tw, "%s\t", units.ByteSize(ds.Summary.Capacity))
    		fmt.Fprintf(tw, "%s\t", units.ByteSize(ds.Summary.FreeSpace))
    		fmt.Fprintf(tw, "\n")
    	}
    tw.Flush()
*/
