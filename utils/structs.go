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
	WinAutoattend        string                   `json:"winAutoattend"`
	SecureBoot           bool                     `json:"secureBoot"`
	TPM                  bool                     `json:"tpm"`
	UserData             string                   `json:"userData"`
	VNCPort              int                      `json:"vncPort"`
	Tags                 []map[string]interface{} `json:"tags"`
	CreatedAt            string                   `json:"createdAt"`
}

type Devices struct {
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces"`
	StorageDisks      []StorageDisk      `json:"storageDisks"`
	CDROMs            []CDROM            `json:"cdroms"`
	FloppyDisks       []FloppyDisk       `json:"floppyDisks"`
}

type FloppyDisk struct {
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Connected   bool   `json:"connected"`
	Path        string `json:"path"`
}

type CDROM struct {
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Connected   bool   `json:"connected"`
	Path        string `json:"path"`
}

type StorageDisk struct {
	IndexNumber  int    `json:"indexNumber"`
	BootOrder    int    `json:"bootOrder"`
	SizeGB       int    `json:"sizeGB"`
	BusType      string `json:"busType"`
	Path         string `json:"path"`
	DatastoreId  string `json:"datastoreId"`
	ExistingPath string `json:"existingPath"`
	Clone        bool   `json:"clone"`
}

type NetworkInterface struct {
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Model       string `json:"model"`
	Connected   bool   `json:"connected"`
	MacAddress  string `json:"macAddress"`
	BridgeName  string `json:"bridgeName"`
	VPCId       string `json:"vpcId"`
}
