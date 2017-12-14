package proxmox

// NextIDResponse is the next available VMID from the Proxmox API
type NextIDResponse struct {
	Data string `json:"data"`
}

// Service contains the methods that the proxmox client provides
type Service interface {
	SignIn() error
	VerifyTicket() (bool, error)
	ResourceList() (Resources, error)

	PickNode() (string, error)

	ContainerCreate(*ContainerCreateRequest) error
	ContainerStop(*ContainerVMStatusRequest) error
	ContainerStart(*ContainerVMStatusRequest) error
	ContainerShutdown(*ContainerVMStatusRequest) error
	ContainerResume(*ContainerVMStatusRequest) error
	ContainerDelete(node string, vmid int) error
	ContainerConfig(*ContainerConfigRequest) (*ContainerConfig, error)

	// VMDelete(node string, vmid int) error
	// VMConfig(*VMConfigRequest) (*VMConfig, error)
	// VMCreate(*VMCreateRequest) error
	// VMStop(*ContainerVMStatusRequest) error
	// VMStart(*ContainerVMStatusRequest) error
	// VMShutdown(*ContainerVMStatusRequest) error
	// VMResume(*ContainerVMStatusRequest) error
	// VMSuspend(*ContainerVMStatusRequest) error
	// VMReset(*ContainerVMStatusRequest) error

	TemplateList(node string) ([]*Template, error)
	ISOList(node string) ([]string, error)

	NextID() (int, error)
}
