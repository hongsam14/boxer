package compose

type Machine struct {
	Image    string `yaml:"image"`
	Snapshot string `yaml:"snapshot"`
	Ip       string `yaml:"ip"`
	OSType   string `yaml:"os_type"`
}
