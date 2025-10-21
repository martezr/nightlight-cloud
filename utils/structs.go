package utils

type InstanceIPMapping struct {
	MacAddress string `json:"macAddress" storm:"id"`
	IPAddress  string `json:"ipAddress" storm:"index"`
}

type Instance struct {
	ID                   string                   `json:"id" storm:"id,index"`
	Name                 string                   `json:"name" storm:"index"`
	Description          string                   `json:"description"`
	InitializationStatus string                   `json:"initializationStatus"`
	BootType             string                   `json:"bootType"`
	CPUCores             int                      `json:"cpuCores"`
	CPUSockets           int                      `json:"cpuSockets"`
	MemoryMB             int                      `json:"memoryMB"`
	PrimaryIPAddress     string                   `json:"primaryIPAddress"`
	PrimaryMacAddress    string                   `json:"primaryMacAddress"`
	MetadataIPAddress    string                   `json:"metadataIPAddress"`
	Devices              Devices                  `json:"devices"`
	PowerState           string                   `json:"powerState"`
	ImageId              string                   `json:"imageId"`
	InstanceProfile      string                   `json:"instanceProfile"`
	DatastoreId          string                   `json:"datastoreId"`
	Kickstart            string                   `json:"kickstart"`
	WinAutoattend        string                   `json:"winAutattend"`
	UserData             string                   `json:"userData"`
	VNCPort              int                      `json:"vncPort"`
	Tags                 []map[string]interface{} `json:"tags"`
}

type Devices struct {
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces"`
	StorageDisks      []StorageDisk      `json:"storageDisks"`
	CDROMs            []CDROM            `json:"cdroms"`
	FloppyDisks       []FloppyDisk       `json:"floppyDisks"`
}

type FloppyDisk struct {
	BootOrder int    `json:"bootOrder"`
	Connected bool   `json:"connected"`
	Path      string `json:"path"`
}

type CDROM struct {
	BootOrder int    `json:"bootOrder"`
	Connected bool   `json:"connected"`
	Path      string `json:"path"`
}

type StorageDisk struct {
	BootOrder    int    `json:"bootOrder"`
	SizeGB       int    `json:"sizeGB"`
	BusType      string `json:"busType"`
	DatastoreId  string `json:"datastoreId"`
	ExistingPath string `json:"existingPath"`
	Clone        bool   `json:"clone"`
}

type NetworkInterface struct {
	BootOrder  int    `json:"bootOrder"`
	Model      string `json:"model"`
	Connected  bool   `json:"connected"`
	MacAddress string `json:"macAddress"`
	BridgeName string `json:"bridgeName"`
}
