package daemon

import (
	"net"

	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
)

const (
	// stockRuntimeName is the reserved name/alias used to represent the
	// OCI runtime being shipped with the docker daemon package.
	stockRuntimeName = "runz"
)

var (
	defaultPidFile = "/system/volatile/docker/docker.pid"
	defaultGraph   = "/var/lib/docker"
	defaultExec    = "zones"
)

// Config defines the configuration of a docker daemon.
// These are the configuration settings that you pass
// to the docker daemon when you launch it with say: `docker -d -e lxc`
type Config struct {
	CommonConfig

	// Fields below here are platform specific.
	ExecRoot       string                   `json:"exec-root,omitempty"`
	ContainerdAddr string                   `json:"containerd,omitempty"`
	Runtimes       map[string]types.Runtime `json:"runtimes,omitempty"`
	DefaultRuntime string                   `json:"default-runtime,omitempty"`
}

// bridgeConfig stores all the bridge driver specific
// configuration.
type bridgeConfig struct {
	commonBridgeConfig

	// Fields below here are platform specific.
	DefaultIP                   net.IP `json:"ip,omitempty"`
	IP                          string `json:"bip,omitempty"`
	DefaultGatewayIPv4          net.IP `json:"default-gateway,omitempty"`
	DefaultGatewayIPv6          net.IP `json:"default-gateway-v6,omitempty"`
	InterContainerCommunication bool   `json:"icc,omitempty"`
}

// InstallFlags adds command-line options to the top-level flag parser for
// the current process.
// Subsequent calls to `flag.Parse` will populate config with values parsed
// from the command-line.
func (config *Config) InstallFlags(cmd *flag.FlagSet, usageFn func(string) string) {
	// First handle install flags which are consistent cross-platform
	config.InstallCommonFlags(cmd, usageFn)

	cmd.StringVar(&config.SocketGroup, []string{"G", "-group"}, "docker", usageFn("Group for the unix socket"))
	cmd.StringVar(&config.bridgeConfig.IP, []string{"#bip", "-bip"}, "", usageFn("Specify network bridge IP"))
	cmd.StringVar(&config.bridgeConfig.Iface, []string{"b", "-bridge"}, "", usageFn("Attach containers to a network bridge"))
	cmd.StringVar(&config.bridgeConfig.FixedCIDR, []string{"-fixed-cidr"}, "", usageFn("IPv4 subnet for fixed IPs"))
	cmd.Var(opts.NewIPOpt(&config.bridgeConfig.DefaultGatewayIPv4, ""), []string{"-default-gateway"}, usageFn("Container default gateway IPv4 address"))
	cmd.Var(opts.NewIPOpt(&config.bridgeConfig.DefaultGatewayIPv6, ""), []string{"-default-gateway-v6"}, usageFn("Container default gateway IPv6 address"))
	cmd.BoolVar(&config.bridgeConfig.InterContainerCommunication, []string{"#icc", "-icc"}, true, usageFn("Enable inter-container communication"))
	cmd.Var(opts.NewIPOpt(&config.bridgeConfig.DefaultIP, "0.0.0.0"), []string{"#ip", "-ip"}, usageFn("Default IP when binding container ports"))
	config.Runtimes = make(map[string]types.Runtime)
	cmd.Var(runconfigopts.NewNamedRuntimeOpt("runtimes", &config.Runtimes, stockRuntimeName), []string{"-add-runtime"}, usageFn("Register an additional OCI compatible runtime"))
	cmd.StringVar(&config.DefaultRuntime, []string{"-default-runtime"}, stockRuntimeName, usageFn("Default OCI runtime to be used"))

	// Then platform-specific install flags
	config.attachExperimentalFlags(cmd, usageFn)
}

// GetRuntime returns the runtime path and arguments for a given
// runtime name
func (config *Config) GetRuntime(name string) *types.Runtime {
	config.reloadLock.Lock()
	defer config.reloadLock.Unlock()
	if rt, ok := config.Runtimes[name]; ok {
		return &rt
	}
	return nil
}

// GetDefaultRuntimeName returns the current default runtime
func (config *Config) GetDefaultRuntimeName() string {
	config.reloadLock.Lock()
	rt := config.DefaultRuntime
	config.reloadLock.Unlock()

	return rt
}

// GetAllRuntimes returns a copy of the runtimes map
func (config *Config) GetAllRuntimes() map[string]types.Runtime {
	config.reloadLock.Lock()
	rts := config.Runtimes
	config.reloadLock.Unlock()
	return rts
}

// GetExecRoot returns the user configured Exec-root
func (config *Config) GetExecRoot() string {
	return config.ExecRoot
}

func (config *Config) isSwarmCompatible() error {
	return nil
}
