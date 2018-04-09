package model

// Volume describes how a local path is mounted into a container.
type Volume struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
	Mode          string `json:"mode"`
}

// Device describes the mapping between a host device and the container device.
type Device struct {
	HostPath          string `json:"host_path"`
	ContainerPath     string `json:"container_path"`
	CgroupPermissions string `json:"cgroup_permissions"`
}

// VolumesFrom describes a container that volumes are imported from.
type VolumesFrom struct {
	Tag           string `json:"tag"`
	Name          string `json:"name"`
	Auth          string `json:"auth"`
	NamePrefix    string `json:"name_prefix"`
	URL           string `json:"url"`
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

// ContainerImage describes a docker container image.
type ContainerImage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
	Auth string `json:"auth"`
	URL  string `json:"url"`
}

// Container describes a container used as part of a DE job.
type Container struct {
	ID             string         `json:"id"`
	Volumes        []Volume       `json:"container_volumes"`
	Devices        []Device       `json:"container_devices"`
	VolumesFrom    []VolumesFrom  `json:"container_volumes_from"`
	Name           string         `json:"name"`
	NetworkMode    string         `json:"network_mode"`
	CPUShares      int64          `json:"cpu_shares"`
	MemoryLimit    int64          `json:"memory_limit"`     // The maximum the container is allowed to have.
	MinMemoryLimit int64          `json:"min_memory_limit"` // The minimum the container needs.
	MinCPUCores    int            `json:"min_cpu_cores"`    // The minimum number of cores the container needs.
	MinDiskSpace   int64          `json:"min_disk_space"`   // The minimum amount of disk space that the container needs.
	PIDsLimit      int64          `json:"pids_limit"`
	Image          ContainerImage `json:"image"`
	EntryPoint     string         `json:"entrypoint"`
	WorkingDir     string         `json:"working_directory"`
}

// WorkingDirectory returns the container's working directory. Defaults to
// /de-app-work if the job submission didn't specify one. Use this function
// rather than accessing the field directly.
func (c *Container) WorkingDirectory() string {
	if c.WorkingDir == "" {
		return "/de-app-work"
	}
	return c.WorkingDir
}

// UsesVolumes returns a boolean value which indicates if a container uses host-mounted volumes
func (c *Container) UsesVolumes() bool {
	if len(c.Volumes) > 0 {
		return true
	}
	return false
}
