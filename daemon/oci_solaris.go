package daemon

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/docker/docker/container"
	"github.com/docker/docker/libcontainerd"
	"github.com/docker/docker/oci"
	containertypes "github.com/docker/engine-api/types/container"
	"github.com/opencontainers/specs/specs-go"
)

func setResources(s *specs.Spec, r containertypes.Resources) error {
	s.Solaris.CappedMemory = getMemoryResources(r)
	s.Solaris.CappedCPU = getCPUResources(r)

	return nil
}

func setUser(s *specs.Spec, c *container.Container) error {
	uid, gid, additionalGids, err := getUser(c, c.Config.User)
	if err != nil {
		return err
	}
	s.Process.User.UID = uid
	s.Process.User.GID = gid
	s.Process.User.AdditionalGids = additionalGids
	return nil
}

func getUser(c *container.Container, username string) (uint32, uint32, []uint32, error) {
	return 0, 0, nil, nil
}

func (daemon *Daemon) populateCommonSpec(s *specs.Spec, c *container.Container) error {
	linkedEnv, err := daemon.setupLinkedContainers(c)
	if err != nil {
		return err
	}
	s.Root = specs.Root{
		Path:     filepath.Join("../../../", filepath.Dir(c.BaseFS)),
		Readonly: c.HostConfig.ReadonlyRootfs,
	}
	rootUID, rootGID := daemon.GetRemappedUIDGID()
	if err := c.SetupWorkingDirectory(rootUID, rootGID); err != nil {
		return err
	}
	cwd := c.Config.WorkingDir
	s.Process.Args = append([]string{c.Path}, c.Args...)
	s.Process.Cwd = cwd
	s.Process.Env = c.CreateDaemonEnvironment(linkedEnv)
	s.Process.Terminal = c.Config.Tty
	s.Hostname = c.FullHostname()

	return nil
}

func (daemon *Daemon) createSpec(c *container.Container) (*libcontainerd.Spec, error) {
	s := oci.DefaultSpec()
	if err := daemon.populateCommonSpec(&s, c); err != nil {
		return nil, err
	}

	if err := setResources(&s, c.HostConfig.Resources); err != nil {
		return nil, fmt.Errorf("runtime spec resources: %v", err)
	}

	if err := setUser(&s, c); err != nil {
		return nil, fmt.Errorf("spec user: %v", err)
	}

	if err := daemon.setupIpcDirs(c); err != nil {
		return nil, err
	}

	ms, err := daemon.setupMounts(c)
	if err != nil {
		return nil, err
	}
	ms = append(ms, c.IpcMounts()...)
	ms = append(ms, c.TmpfsMounts()...)
	sort.Sort(mounts(ms))

	return (*libcontainerd.Spec)(&s), nil
}
