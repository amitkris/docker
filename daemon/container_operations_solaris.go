// +build solaris

package daemon

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/container"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/mount"
	networktypes "github.com/docker/engine-api/types/network"
	"github.com/docker/libnetwork"
)

func (daemon *Daemon) setupLinkedContainers(container *container.Container) ([]string, error) {
	return nil, nil
}

// ConnectToNetwork connects a container to a network
func (daemon *Daemon) ConnectToNetwork(container *container.Container, idOrName string, endpointConfig *networktypes.EndpointSettings) error {
	return nil
}

// getSize returns real size & virtual size
func (daemon *Daemon) getSize(container *container.Container) (int64, int64) {
	var (
		sizeRw, sizeRootfs int64
		err                error
	)

	if err := daemon.Mount(container); err != nil {
		logrus.Errorf("Failed to compute size of container rootfs %s: %s", container.ID, err)
		return sizeRw, sizeRootfs
	}
	defer daemon.Unmount(container)

	sizeRw, err = container.RWLayer.Size()
	if err != nil {
		logrus.Errorf("Driver %s couldn't return diff size of container %s: %s",
			daemon.GraphDriverName(), container.ID, err)
		// FIXME: GetSize should return an error. Not changing it now in case
		// there is a side-effect.
		sizeRw = -1
	}

	if parent := container.RWLayer.Parent(); parent != nil {
		sizeRootfs, err = parent.Size()
		if err != nil {
			sizeRootfs = -1
		} else if sizeRw != -1 {
			sizeRootfs += sizeRw
		}
	}
	return sizeRw, sizeRootfs
}

// DisconnectFromNetwork disconnects a container from the network
func (daemon *Daemon) DisconnectFromNetwork(container *container.Container, n libnetwork.Network, force bool) error {
	return nil
}

func (daemon *Daemon) setupIpcDirs(container *container.Container) error {
	return nil
}

func (daemon *Daemon) mountVolumes(container *container.Container) error {
	mounts, err := daemon.setupMounts(container)
	if err != nil {
		return err
	}

	for _, m := range mounts {
		dest, err := container.GetResourcePath(m.Destination)
		if err != nil {
			return err
		}

		var stat os.FileInfo
		stat, err = os.Stat(m.Source)
		if err != nil {
			return err
		}
		if err = fileutils.CreateIfNotExists(dest, stat.IsDir()); err != nil {
			return err
		}

		opts := "rbind,ro"
		if m.Writable {
			opts = "rbind,rw"
		}

		if err := mount.Mount(m.Source, dest, "lofs", opts); err != nil {
			return err
		}
	}

	return nil
}

func killProcessDirectly(container *container.Container) error {
	return nil
}

func detachMounted(path string) error {
	return nil
}

func isLinkable(child *container.Container) bool {
	// A container is linkable only if it belongs to the default network
	_, ok := child.NetworkSettings.Networks["bridge"]
	return ok
}

func errRemovalContainer(containerID string) error {
	return fmt.Errorf("Container %s is marked for removal and cannot be connected or disconnected to the network", containerID)
}
