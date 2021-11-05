package main

type KVMNodeList struct {
	Name string `json:"name"`
	Id   uint   `json:"id"`
}

type VirDiskSource struct {
	File string `xml:"file,attr"`
}
type VirDiskTarget struct {
	Dev string `xml:"dev,attr"`
}

type VirDisk struct {
	Type   string        `xml:"type,attr"`
	Source VirDiskSource `xml:"source"`
	Target VirDiskTarget `xml:"target"`
}

type VirDevice struct {
	Disks []VirDisk `xml:"disk"`
}

type VirDomain struct {
	Name    string    `xml:"name"`
	UUID    string    `xml:"uuid"`
	Memory  string    `xml:"memory"`
	Devices VirDevice `xml:"devices"`
}

type MasterVM struct {
	MasterVMName string `json:"mastervm"`
}
