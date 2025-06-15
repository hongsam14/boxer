# Boxer
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

Boxer is a program that aims to common interface to control VM software that runs on cli, such as VirtualBox and Kvm.
Users can control the VMs written in the config using the boxer client interface.

![DALL·E 2025-01-08 23 37 06 - A c](https://github.com/user-attachments/assets/0ef3f3e1-63de-4ac8-8eff-fcc8586d8dc3)

## Key Concept: Box

Boxer groups multiple VMs together and handles the VMs as an abstract object called a box.

``` Go
type VMInfoConfig struct {
	// Name is the name of the VM
	Name string `mapstructure:"name" yaml:"name"`
	// Snapshot is the name of the snapshot
	Snapshot string `mapstructure:"snapshot" yaml:"snapshot"`
	IP       string `mapstructure:"ip" yaml:"ip"`       // IP is the IP address of the VM
	OS       string `mapstructure:"os" yaml:"os"`       // OS is the operating system of the VM
	Group    string `mapstructure:"group" yaml:"group"` // Group is the group of the VM, used for grouping VMs in the UI
}
```
Users can specify a `Group` when defining the VMs to use in config.

``` Go
// Example code
//
VMInfo: map[string]config.VMInfoConfig{
		"sb_win10_develop_v2": {
			Name:     "sb_win10_develop_v2",
			Snapshot: "snapshot0",
			OS:       "windows",
			Group:    "testGroup",
			IP:       "127.0.0.1",
		},
		"sb_win10_develop_v2_clone_0": {
			Name:     "sb_win10_develop_v2_clone_0",
			Snapshot: "snapshot0",
			OS:       "windows",
			Group:    "testGroup",
			IP:       "127.0.0.2",
		},
		"openssh": {
			Name:     "openssh",
			Snapshot: "Snapshot 1",
			OS:       "linux",
			Group:    "testGroup2",
			IP:       "127.0.0.3",
		},
	},
  ...
```
In the example above, `sb_win10_develop_v2` and `sb_win10_develop_v2_clone_0` are defined as the same group `testGroup`.

A user can be assigned a vm that is available as a group argument.
``` Go
  // allocate a Box for the test group
	box, err := client.Balloc("testGroup2")
	if err != nil {
		t.Fatalf("Failed to allocate Box: %v", err)
		return
	}
  // deallocate the Box
	err = client.Bfree(box)
	if err != nil {
		t.Fatalf("Failed to deallocate Box: %v", err)
		return
	}
```
A user can be assigned a vm that is available as a group argument. And user can free the `box` after you are done using `Bfree`.
This way, users can be assigned any available `box` in the group without worrying about the name of the VM. This approach can be useful when managing many VMs for different purposes.

## Key Concept: Just 3 vm operations

Boxer supports only three VM operation:
- Start   : Start vm
- Stop    : Stop vm
- Restore : Restore vm by snapshot

Boxer does not support any other VM operations (creation, deletion, editing, communication, etc.) and only supports VM control, thus avoiding dependency on a specific VM vendor.
Boxer simply leverages existing VMs and does not make any changes to the VM environment.

This intention is implemented by allowing users to define control configs.

``` Go
  // when using VirtualBox

  VMControl: config.VMControlConfig{
		StartCmd:           "VBoxManage startvm $machine",
		StopCmd:            "VBoxManage controlvm $machine poweroff",
		RestoreSnapshotCmd: "VBoxManage snapshot $machine restore $snapshot",
	},

  // when using kvm
  VMControl: config.VMControlConfig{
		StartCmd:           "virsh start $machine",
		StopCmd:            "virsh shutdown $machine",
		RestoreSnapshotCmd: "virsh snapshot-revert $machine $snapshot",
	},
  ...
```
You can set vm control commands in config using the reserved words $machine, $snapshot.

## Future plans & usage

Boxer is expected to be used to develop applications that need to control sandbox-like VMs.
