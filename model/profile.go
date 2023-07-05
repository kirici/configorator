package model

type Profile struct {
	Config                Config                `json:"config"`
	NotificationEndpoints NotificationEndpoints `json:"notification_endpoints"`
}
type Config struct {
	AllowWorkloadsOnMaster  bool               `json:"allowWorkloadsOnMaster"`
	BaseDomain              string             `json:"baseDomain"`
	BaseIPType              string             `json:"baseIpType"`
	BasePassword            string             `json:"basePassword"`
	BaseUser                string             `json:"baseUser"`
	CertificationMode       bool               `json:"certificationMode"`
	DefaultKeyboardLanguage string             `json:"defaultKeyboardLanguage"`
	DisableLinuxDesktop     bool               `json:"disableLinuxDesktop"`
	DisableSessionTimeout   bool               `json:"disableSessionTimeout"`
	DNSResolution           string             `json:"dnsResolution"`
	Docker                  Docker             `json:"docker"`
	EnvironmentPrefix       string             `json:"environmentPrefix"`
	GlusterFsDiskSize       int                `json:"glusterFsDiskSize"`
	KubeOrchestrator        string             `json:"kubeOrchestrator"`
	LocalVolumes            LocalVolumes       `json:"local_volumes"`
	LogLevel                string             `json:"logLevel"`
	MetalLbIPRange          MetalLBIPRange     `json:"metalLbIpRange"`
	ProxySettings           ProxySettings      `json:"proxy_settings"`
	SelectedTemplates       string             `json:"selectedTemplates"`
	SslProvider             string             `json:"sslProvider"`
	StandaloneMode          bool               `json:"standaloneMode"`
	StartupMode             string             `json:"startupMode"`
	StaticNetworkSetup      StaticNetworkSetup `json:"staticNetworkSetup"`
	UpdateSourceOnStart     bool               `json:"updateSourceOnStart"`
	VirtualizationType      string             `json:"virtualizationType"`
	VMProperties            VMProperties       `json:"vm_properties"`
}
type Docker struct {
	DockerhubEmail    string `json:"dockerhub_email"`
	DockerhubPassword string `json:"dockerhub_password"`
	DockerhubUsername string `json:"dockerhub_username"`
}
type LocalVolumes struct {
	FiftyGB  int `json:"fifty_gb"`
	FiveGB   int `json:"five_gb"`
	OneGB    int `json:"one_gb"`
	TenGB    int `json:"ten_gb"`
	ThirtyGB int `json:"thirty_gb"`
}

type MetalLBIPRange struct {
	IPRangeEnd   string `json:"ipRangeEnd"`
	IPRangeStart string `json:"ipRangeStart"`
}
type ProxySettings struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
	NoProxy    string `json:"no_proxy"`
}

type StaticNetworkSetup struct {
	BaseFixedIPAddresses BaseFixedIPAddresses `json:"baseFixedIpAddresses"`
	Dns1                 string               `json:"dns1"`
	Dns2                 string               `json:"dns2"`
	Gateway              string               `json:"gateway"`
}

type BaseFixedIPAddresses struct {
	KxMain1   string `json:"kx-main1"`
	KxMain2   string `json:"kx-main2"`
	KxMain3   string `json:"kx-main3"`
	KxWorker1 string `json:"kx-worker1"`
	KxWorker2 string `json:"kx-worker2"`
	KxWorker3 string `json:"kx-worker3"`
	KxWorker4 string `json:"kx-worker4"`
}

type VMProperties struct {
	ThreeDAcceleration      string `json:"3d_acceleration"`
	MainAdminNodeCPUCores   int    `json:"main_admin_node_cpu_cores"`
	MainAdminNodeMemory     int    `json:"main_admin_node_memory"`
	MainNodeCount           int    `json:"main_node_count"`
	MainReplicaNodeCPUCores int    `json:"main_replica_node_cpu_cores"`
	MainReplicaNodeMemory   int    `json:"main_replica_node_memory"`
	WorkerNodeCount         int    `json:"worker_node_count"`
	WorkerNodeCPUCores      int    `json:"worker_node_cpu_cores"`
	WorkerNodeMemory        int    `json:"worker_node_memory"`
}
type NotificationEndpoints struct {
	EmailAddress   string `json:"email_address"`
	MsTeamsWebhook string `json:"ms_teams_webhook"`
	SlackWebhook   string `json:"slack_webhook"`
}
