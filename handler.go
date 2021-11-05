package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/libvirt/libvirt-go"
)

func GetKVMLists(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// w.WriteHeader(http.StatusOK)

	uri := "qemu:///system"
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		fmt.Println("failed to connect to qemu")
		fmt.Println(err)
	}
	// fmt.Println(conn)
	defer conn.Close()
	// libvirt.CONNECT_LIST_DOMAINS_ACTIVE
	// libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	// libvirt.CONNECT_LIST_DOMAINS_PERSISTENT
	lists, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_PERSISTENT)
	if err != nil {
		fmt.Println("failed to get Lists of All Domains")
		json.NewEncoder(w).Encode(err)

	}
	var nodelists []KVMNodeList
	for _, dom := range lists {
		name, _ := dom.GetName()
		id, _ := dom.GetID()
		nodelists = append(nodelists, KVMNodeList{name, id})
		dom.Free()
	}
	fmt.Println(nodelists)
	json.NewEncoder(w).Encode(nodelists)

}

func StartNode(w http.ResponseWriter, r *http.Request) {

	vmName := r.URL.Query().Get("node")
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Println("connection error: ", err)
		errJson := jsonErr{500, "connection error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	defer conn.Close()
	vm, err := conn.LookupDomainByName(vmName)
	if err != nil {
		fmt.Println("lookupByName Error: ", err)
		errJson := jsonErr{500, "lookupByName Error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}

	if err := vm.Create(); err != nil {
		fmt.Println("StartVM Error: ", err)
		errJson := jsonErr{500, "StartVM Error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}

	defer vm.Free()
	successJson := jsonErr{200, "Success", "StartVM Success"}
	json.NewEncoder(w).Encode(successJson)
}

func StopNode(w http.ResponseWriter, r *http.Request) {

	vmName := r.URL.Query().Get("node")
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Println("connection error: ", err)
		errJson := jsonErr{500, "connection error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	defer conn.Close()
	vm, err := conn.LookupDomainByName(vmName)
	if err != nil {
		fmt.Println("lookupByName Error: ", err)
		errJson := jsonErr{500, "lookupByName Error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}

	if err := vm.Shutdown(); err != nil {
		fmt.Println("StopVM Error: ", err)
		errJson := jsonErr{500, "StopVM Error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}

	defer vm.Free()
	successJson := jsonErr{200, "Success", "StopVM Success"}
	json.NewEncoder(w).Encode(successJson)
}

func ChangeNode(w http.ResponseWriter, r *http.Request) {
	vmName := r.URL.Query().Get("node")
	cpu := r.URL.Query().Get("cpu")
	memory := r.URL.Query().Get("mem")

	fmt.Println(vmName, cpu, memory)
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Println("connection error: ", err)
		errJson := jsonErr{500, "connection error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	defer conn.Close()
	targetVM, err := conn.LookupDomainByName(vmName)
	if err != nil {
		fmt.Println("lookupByName Error: ", err)
		errJson := jsonErr{500, "lookupByName Error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}

	fmt.Println(targetVM.GetMaxVcpus())
	fmt.Println(targetVM.GetVcpus())
	fmt.Println(targetVM.GetMaxMemory())
	if cpu != "" {
		vcpu, err := strconv.ParseUint(cpu, 10, 32)
		var vErr error
		flags := libvirt.DOMAIN_VCPU_MAXIMUM | libvirt.DomainVcpuFlags(libvirt.DOMAIN_AFFECT_CONFIG)
		vErr = targetVM.SetVcpusFlags(uint(vcpu), flags)
		if vErr != nil {
			fmt.Println(vErr)
			errJson := jsonErr{500, "Maximun vCPUs change Error", err.Error()}
			json.NewEncoder(w).Encode(errJson)
		}
		vErr = targetVM.SetVcpusFlags(uint(vcpu), libvirt.DOMAIN_VCPU_CONFIG)
		if vErr != nil {
			fmt.Println(vErr)
			errJson := jsonErr{500, "config vCPUs change Error", err.Error()}
			json.NewEncoder(w).Encode(errJson)
		}
		if memory != "" {
			mem, _ := strconv.ParseUint(memory, 10, 32)
			mem = mem * 1024

			memflags := libvirt.DOMAIN_MEM_MAXIMUM | libvirt.DomainMemoryModFlags(libvirt.DOMAIN_AFFECT_CONFIG)
			vErr = targetVM.SetMemoryFlags(uint64(mem), memflags)
			if vErr != nil {
				fmt.Println("create fail 1")
				fmt.Println(vErr)
			}

			vErr = targetVM.SetMemoryFlags(uint64(mem), libvirt.DOMAIN_MEM_CONFIG)
			if vErr != nil {
				fmt.Println("create fail 1")
				fmt.Println(vErr)
			}
		}

		// shutdown Domain
		vmState, _, _ := targetVM.GetState()
		if vmState == libvirt.DOMAIN_RUNNING {
			vErr = targetVM.ShutdownFlags(libvirt.DOMAIN_SHUTDOWN_DEFAULT)
			if vErr != nil {
				fmt.Println("shutdown fail")
				errJson := jsonErr{500, "shutdown fail", err.Error()}
				json.NewEncoder(w).Encode(errJson)
			}
		}

		for i := 1; i <= 30; i++ {
			// VIR_DOMAIN_NOSTATE	=	0 (0x0)	no state
			// VIR_DOMAIN_RUNNING	=	1 (0x1)	the domain is running
			// VIR_DOMAIN_BLOCKED	=	2 (0x2)	the domain is blocked on resource
			// VIR_DOMAIN_PAUSED	=	3 (0x3)	the domain is paused by user
			// VIR_DOMAIN_SHUTDOWN	=	4 (0x4)	the domain is being shut down
			// VIR_DOMAIN_SHUTOFF	=	5 (0x5)	the domain is shut off
			// VIR_DOMAIN_CRASHED	=	6 (0x6)	the domain is crashed
			// VIR_DOMAIN_PMSUSPENDED	=	7 (0x7)	the domain is suspended by guest power management
			// VIR_DOMAIN_LAST	=	8 (0x8)	NB: this enum value will increase over time as new events are added to the libvirt API. It reflects the last state supported by this version of the libvirt API.

			vmState, _, _ = targetVM.GetState()
			fmt.Println(vmState)
			if vmState == libvirt.DOMAIN_SHUTOFF {
				// time.Sleep(10 * time.Second)
				break
			}
			time.Sleep(1 * time.Second)
		}

		// start Domain
		vErr = targetVM.Create()
		if vErr != nil {
			fmt.Println("start fail")
			errJson := jsonErr{500, "start fail", err.Error()}
			json.NewEncoder(w).Encode(errJson)
		}

		defer targetVM.Free()
		successJson := jsonErr{200, "success", "node spec changed"}
		json.NewEncoder(w).Encode(successJson)
	}
}

func CreateNode(w http.ResponseWriter, r *http.Request) {
	templateVM := r.URL.Query().Get("template")
	newVM := r.URL.Query().Get("newvm")
	masterName := r.URL.Query().Get("master")
	masterPass := r.URL.Query().Get("mpass")
	workerPass := r.URL.Query().Get("wpass")
	curi := "qemu:///system"

	virtClone, err := exec.Command("bash", "-c", "which virt-clone").Output()
	if err != nil {
		fmt.Println("find virt-clone failed")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	whichStr := strings.Trim(string(virtClone), "\n")

	cmd := fmt.Sprintf("%s --connect %s -o %s -n %s --auto-clone", whichStr, curi, templateVM, newVM)
	fmt.Println(cmd)

	fcmd := exec.Command("bash", "-c", cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	fcmd.Stdout = &out
	fcmd.Stderr = &stderr
	err = fcmd.Start()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	err = fcmd.Wait()
	if err != nil {
		fmt.Println("clone fail")
		errJson := jsonErr{500, "fail", stderr.String()}
		json.NewEncoder(w).Encode(errJson)
	} else {
		conn, err := libvirt.NewConnect("qemu:///system")
		if err != nil {
			fmt.Println("connection error: ", err)
			errJson := jsonErr{500, "connection error", err.Error()}
			json.NewEncoder(w).Encode(errJson)
		}
		newVMIns, _ := conn.LookupDomainByName(newVM)
		defer newVMIns.Free()
		err = newVMIns.Create()
		if err != nil {
			fmt.Println(err)
			fmt.Println("clone fail")
			errJson := jsonErr{500, "fail", stderr.String()}
			json.NewEncoder(w).Encode(errJson)
		}
	}
	successJson := jsonErr{200, "success", "node clone success"}
	json.NewEncoder(w).Encode(successJson)

	go func() {
		NodeJoin(masterName, newVM, masterPass, workerPass)
	}()
}

func DeleteNode(w http.ResponseWriter, r *http.Request) {
	vmName := r.URL.Query().Get("node")

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Println("connection error: ", err)
		errJson := jsonErr{500, "connection error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	defer conn.Close()

	deleteTarget, err := conn.LookupDomainByName(vmName)
	if err != nil {
		fmt.Println("domain search fail")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
		return
	}

	xmlBytes, _ := deleteTarget.GetXMLDesc(0)

	var data VirDomain
	var sourceFile string
	err = xml.Unmarshal([]byte(xmlBytes), &data)
	if err != nil {
		fmt.Println("xml parse error")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
		return
	}
	for _, blk := range data.Devices.Disks {
		if blk.Target.Dev == "vda" {
			sourceFile = blk.Source.File
			break
		}
	}

	vmState, _, _ := deleteTarget.GetState()
	if vmState != libvirt.DOMAIN_SHUTOFF {
		err = deleteTarget.Destroy()
		if err != nil {
			fmt.Println("vm destory fail")
			errJson := jsonErr{500, "fail", err.Error()}
			json.NewEncoder(w).Encode(errJson)
			return
		}
	}
	err = deleteTarget.Undefine()
	if err != nil {
		fmt.Println("vm undefine fail")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
		return
	}

	tagetVol, err := conn.LookupStorageVolByPath(sourceFile)
	if err != nil {
		fmt.Println("find vol fail")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
		return
	}

	err = tagetVol.Delete(libvirt.STORAGE_VOL_DELETE_NORMAL)
	if err != nil {
		fmt.Println("Delete vol fail")
		errJson := jsonErr{500, "fail", err.Error()}
		json.NewEncoder(w).Encode(errJson)
		return
	}

	successJson := jsonErr{200, "success", "node deleted"}
	json.NewEncoder(w).Encode(successJson)
}

func GetMasterVMName(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Println("connection error: ", err)
		errJson := jsonErr{500, "connection error", err.Error()}
		json.NewEncoder(w).Encode(errJson)
	}
	defer conn.Close()

	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		fmt.Println(err)
	}
	masterVM := ""
	for _, dom := range domains {
		// hn, err := dom.GetHostname(libvirt.DOMAIN_GET_HOSTNAME_LEASE)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println(hn)

		desc, err := dom.GetMetadata(libvirt.DOMAIN_METADATA_DESCRIPTION, "", 0)
		if err != nil {
			// fmt.Println(err)
		} else {
			fmt.Println(desc)
			if strings.TrimSpace(desc) == hostname {
				masterVM, _ = dom.GetName()
				// fmt.Println("masterVM=", masterVM)
				break
			}
		}
	}
	masterJson := MasterVM{masterVM}
	fmt.Println(masterVM)
	json.NewEncoder(w).Encode(masterJson)

}
