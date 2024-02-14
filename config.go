package faraday

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/lightninglabs/faraday/chain"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/cert"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultRPCPort        = "10009"
	defaultRPCHostPort    = "localhost:" + defaultRPCPort
	DefaultNetwork        = "mainnet"
	defaultMinimumMonitor = time.Hour * 24 * 7 * 4 // four weeks in hours
	defaultDebugLevel     = "info"
	defaultRPCListen      = "localhost:8465"

	// By default we do not require connecting to a bitcoin node so that
	// we can serve basic functionality by default.
	defaultChainConn = false

	// defaultTLSCertDuration is the default validity of a self-signed
	// certificate. The value corresponds to 14 months
	// (14 months * 30 days * 24 hours).
	defaultTLSCertDuration = 14 * 30 * 24 * time.Hour
)

var (
	// FaradayDirBase is the default main directory where faraday stores its
	// data.
	FaradayDirBase = btcutil.AppDataDir("faraday", false)

	// DefaultTLSCertFilename is the default file name for the autogenerated
	// TLS certificate.
	DefaultTLSCertFilename = "tls.cert"

	// DefaultTLSKeyFilename is the default file name for the autogenerated
	// TLS key.
	DefaultTLSKeyFilename = "tls.key"

	defaultSelfSignedOrganization = "faraday autogenerated cert"

	// DefaultTLSCertPath is the default full path of the autogenerated TLS
	// certificate.
	DefaultTLSCertPath = filepath.Join(
		FaradayDirBase, DefaultNetwork, DefaultTLSCertFilename,
	)

	// DefaultTLSKeyPath is the default full path of the autogenerated TLS
	// key.
	DefaultTLSKeyPath = filepath.Join(
		FaradayDirBase, DefaultNetwork, DefaultTLSKeyFilename,
	)

	// DefaultMacaroonFilename is the default file name for the
	// autogenerated faraday macaroon.
	DefaultMacaroonFilename = "faraday.macaroon"

	// DefaultMacaroonPath is the default full path of the base faraday
	// macaroon.
	DefaultMacaroonPath = filepath.Join(
		FaradayDirBase, DefaultNetwork, DefaultMacaroonFilename,
	)

	// defaultLndMacaroon is the default macaroon file we use to connect to
	// lnd.
	defaultLndMacaroon = "admin.macaroon"

	// DefaultLndDir is the default location where we look for lnd's tls and
	// macaroon files.
	DefaultLndDir = btcutil.AppDataDir("lnd", false)

	// DefaultLndMacaroonPath is the default location where we look for a
	// macaroon to use when connecting to lnd.
	DefaultLndMacaroonPath = filepath.Join(
		DefaultLndDir, "data", "chain", "bitcoin", DefaultNetwork,
		defaultLndMacaroon,
	)
)

type LndConfig struct {
	// RPCServer is host:port that lnd's RPC server is listening on.
	RPCServer string `long:"rpcserver" description:"host:port that LND is listening for RPC connections on"`

	// MacaroonDir is the directory that contains all the macaroon files
	// required for the remote connection.
	MacaroonDir string `long:"macaroondir" description:"DEPRECATED: Use macaroonpath."`

	// MacaroonPath is the path to the single macaroon that should be used
	// instead of needing to specify the macaroon directory that contains
	// all of lnd's macaroons. The specified macaroon MUST have all
	// permissions that all the subservers use, otherwise permission errors
	// will occur.
	MacaroonPath string `long:"macaroonpath" description:"The full path to the single macaroon to use, either the admin.macaroon or a custom baked one. Cannot be specified at the same time as macaroondir. A custom macaroon must contain ALL permissions required for all subservers to work, otherwise permission errors will occur."`

	// TLSCertPath is the path to the tls cert that faraday should use.
	TLSCertPath string `long:"tlscertpath" description:"Path to TLS cert"`

	// RequestTimeout is the maximum time to wait for a response from lnd.
	RequestTimeout time.Duration `long:"requesttimeout" description:"The maximum time to wait for a response from lnd, if not set the default of 30 seconds will be used."`
}

type Config struct { //nolint:maligned
	// Lnd holds the configuration options for the connection to lnd.
	Lnd *LndConfig `group:"lnd" namespace:"lnd"`

	// FaradayDir is the main directory where faraday stores all its data.
	FaradayDir string `long:"faradaydir" description:"The directory for all of faraday's data. If set, this option overwrites --macaroonpath, --tlscertpath and --tlskeypath."`

	// ChainConn specifies whether to attempt connecting to a bitcoin backend.
	ChainConn bool `long:"connect_bitcoin" description:"Whether to attempt to connect to a backing bitcoin node. Some endpoints will not be available if this option is not enabled."`

	ShowVersion bool `long:"version" description:"Display version information and exit"`

	// MinimumMonitored is the minimum amount of time that a channel must be monitored for before we consider it for termination.
	MinimumMonitored time.Duration `long:"min_monitored" description:"The minimum amount of time that a channel must be monitored for before recommending termination. Valid time units are {s, m, h}."`

	// Network is a string containing the network we're running on.
	Network string `long:"network" description:"The network to run on." choice:"regtest" choice:"testnet" choice:"mainnet" choice:"simnet"`

	// DebugLevel is a string defining the log level for the service either
	// for all subsystems the same or individual level by subsystem.
	DebugLevel string `long:"debuglevel" description:"Debug level for faraday and its subsystems."`

	TLSCertPath        string        `long:"tlscertpath" description:"Path to write the TLS certificate for faraday's RPC and REST services."`
	TLSKeyPath         string        `long:"tlskeypath" description:"Path to write the TLS private key for faraday's RPC and REST services."`
	TLSExtraIPs        []string      `long:"tlsextraip" description:"Adds an extra IP to the generated certificate."`
	TLSExtraDomains    []string      `long:"tlsextradomain" description:"Adds an extra domain to the generated certificate."`
	TLSAutoRefresh     bool          `long:"tlsautorefresh" description:"Re-generate TLS certificate and key if the IPs or domains are changed."`
	TLSDisableAutofill bool          `long:"tlsdisableautofill" description:"Do not include the interface IPs or the system hostname in TLS certificate, use first --tlsextradomain as Common Name instead, if set."`
	TLSCertDuration    time.Duration `long:"tlscertduration" description:"The duration for which the auto-generated TLS certificate will be valid for."`

	MacaroonPath string `long:"macaroonpath" description:"Path to write the macaroon for faraday's RPC and REST services if it doesn't exist."`

	// RPCListen is the listen address for the faraday rpc server.
	RPCListen string `long:"rpclisten" description:"Address to listen on for gRPC clients."`

	// RESTListen is the listen address for the faraday REST server.
	RESTListen string `long:"restlisten" description:"Address to listen on for REST clients. If not specified, no REST listener will be started."`

	// CORSOrigin specifies the CORS header that should be set on REST responses. No header is added if the value is empty.
	CORSOrigin string `long:"corsorigin" description:"The value to send in the Access-Control-Allow-Origin header. Header will be omitted if empty."`

	// Bitcoin is the configuration required to connect to a bitcoin node.
	Bitcoin *chain.BitcoinConfig `group:"bitcoin" namespace:"bitcoin"`
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Lnd: &LndConfig{
			RPCServer:    defaultRPCHostPort,
			MacaroonPath: DefaultLndMacaroonPath,
		},
		FaradayDir:       FaradayDirBase,
		Network:          DefaultNetwork,
		MinimumMonitored: defaultMinimumMonitor,
		DebugLevel:       defaultDebugLevel,
		TLSCertPath:      DefaultTLSCertPath,
		TLSKeyPath:       DefaultTLSKeyPath,
		TLSCertDuration:  defaultTLSCertDuration,
		MacaroonPath:     DefaultMacaroonPath,
		RPCListen:        defaultRPCListen,
		ChainConn:        defaultChainConn,
		Bitcoin:          chain.DefaultConfig,
	}
}

// ValidateConfig sanitizes all file system paths and makes sure no incompatible
// configuration combinations are used.
func ValidateConfig(config *Config) error {
	// Validate the network.
	_, err := lndclient.Network(config.Network).ChainParams()
	if err != nil {
		return fmt.Errorf("error validating network: %v", err)
	}

	// Clean up and validate paths, then make sure the directories exist.
	config.FaradayDir = lncfg.CleanAndExpandPath(config.FaradayDir)
	config.TLSCertPath = lncfg.CleanAndExpandPath(config.TLSCertPath)
	config.TLSKeyPath = lncfg.CleanAndExpandPath(config.TLSKeyPath)
	config.MacaroonPath = lncfg.CleanAndExpandPath(config.MacaroonPath)

	// Append the network type to faraday directory so they are "namespaced"
	// per network.
	config.FaradayDir = filepath.Join(config.FaradayDir, config.Network)

	// Create the full path of directories now, including the network path.
	if err := os.MkdirAll(config.FaradayDir, os.ModePerm); err != nil {
		return err
	}

	// Since our faraday directory overrides our TLS dir and macaroon path
	// values, make sure that they are not set when faraday dir is set. We
	// fail hard here rather than overwriting and potentially confusing the
	// user.
	faradayDirSet := config.FaradayDir != FaradayDirBase
	if faradayDirSet {
		tlsCertPathSet := config.TLSCertPath != DefaultTLSCertPath
		tlsKeyPathSet := config.TLSKeyPath != DefaultTLSKeyPath
		macaroonPathSet := config.MacaroonPath != DefaultMacaroonPath

		if tlsCertPathSet {
			return fmt.Errorf("faradaydir overwrites " +
				"tlscertpath, please only set one value")
		}

		if tlsKeyPathSet {
			return fmt.Errorf("faradaydir overwrites " +
				"tlskeypath, please only set one value")
		}

		if macaroonPathSet {
			return fmt.Errorf("faradaydir overwrites " +
				"macaroonpath, please only set one value")
		}
	}

	// We want the TLS files to also be in the "namespaced" sub directory.
	// Replace the default values with actual values in case the user
	// specified faradaydir.
	if config.TLSCertPath == DefaultTLSCertPath {
		config.TLSCertPath = filepath.Join(
			config.FaradayDir, DefaultTLSCertFilename,
		)
	}
	if config.TLSKeyPath == DefaultTLSKeyPath {
		config.TLSKeyPath = filepath.Join(
			config.FaradayDir, DefaultTLSKeyFilename,
		)
	}
	if config.MacaroonPath == DefaultMacaroonPath {
		config.MacaroonPath = filepath.Join(
			config.FaradayDir, DefaultMacaroonFilename,
		)
	}

	// If the user has opted into connecting to a bitcoin backend, check
	// that we have a rpc user and password, and that tls path is set if
	// required.
	if config.ChainConn {
		if config.Bitcoin.User == "" || config.Bitcoin.Password == "" {
			return fmt.Errorf("rpc user and password " +
				"required when chainconn is set")
		}

		if config.Bitcoin.UseTLS && config.Bitcoin.TLSPath == "" {
			return fmt.Errorf("bitcoin.tlspath required " +
				"when chainconn is set")
		}
	}

	// Make sure only one of the macaroon options is used.
	switch {
	case config.Lnd.MacaroonPath != DefaultLndMacaroonPath &&
		config.Lnd.MacaroonDir != "":

		return fmt.Errorf("use --lnd.macaroonpath only")

	case config.Lnd.MacaroonDir != "":
		// With the new version of lndclient we can only specify a
		// single macaroon instead of all of them. If the old
		// macaroondir is used, we use the readonly macaroon located in
		// that directory.
		config.Lnd.MacaroonPath = path.Join(
			lncfg.CleanAndExpandPath(config.Lnd.MacaroonDir),
			defaultLndMacaroon,
		)

	case config.Lnd.MacaroonPath != "":
		config.Lnd.MacaroonPath = lncfg.CleanAndExpandPath(
			config.Lnd.MacaroonPath,
		)

	default:
		return fmt.Errorf("must specify --lnd.macaroonpath")
	}

	// Adjust the default lnd macaroon path if only the network is
	// specified.
	if config.Network != DefaultNetwork &&
		config.Lnd.MacaroonPath == DefaultLndMacaroonPath {

		config.Lnd.MacaroonPath = path.Join(
			DefaultLndDir, "data", "chain", "bitcoin",
			config.Network, defaultLndMacaroon,
		)
	}

	// Expand the lnd cert path, in case the user is specifying the home
	// directory with ~ (which only the shell understands, not the low-level
	// file system).
	config.Lnd.TLSCertPath = lncfg.CleanAndExpandPath(
		config.Lnd.TLSCertPath,
	)

	return nil
}

// getTLSConfig generates a new self signed certificate or refreshes an existing
// one if necessary, then returns the full TLS configuration for initializing
// a secure server interface.
func getTLSConfig(cfg *Config) (*tls.Config, *credentials.TransportCredentials,
	error) {

	// Let's load our certificate first or create then load if it doesn't
	// yet exist.
	certData, parsedCert, err := loadCertWithCreate(cfg)
	if err != nil {
		return nil, nil, err
	}

	// If the certificate expired or it was outdated, delete it and the TLS
	// key and generate a new pair.
	if time.Now().After(parsedCert.NotAfter) {
		log.Info("TLS certificate is expired or outdated, " +
			"removing old file then generating a new one")

		err := os.Remove(cfg.TLSCertPath)
		if err != nil {
			return nil, nil, err
		}

		err = os.Remove(cfg.TLSKeyPath)
		if err != nil {
			return nil, nil, err
		}

		certData, _, err = loadCertWithCreate(cfg)
		if err != nil {
			return nil, nil, err
		}
	}

	tlsCfg := cert.TLSConfFromCert(certData)
	restCreds, err := credentials.NewClientTLSFromFile(
		cfg.TLSCertPath, "",
	)
	if err != nil {
		return nil, nil, err
	}

	return tlsCfg, &restCreds, nil
}

// loadCertWithCreate tries to load the TLS certificate from disk. If the
// specified cert and key files don't exist, the certificate/key pair is created
// first.
func loadCertWithCreate(cfg *Config) (tls.Certificate, *x509.Certificate,
	error) {

	// Ensure we create TLS key and certificate if they don't exist.
	if !lnrpc.FileExists(cfg.TLSCertPath) &&
		!lnrpc.FileExists(cfg.TLSKeyPath) {

		log.Infof("Generating TLS certificates...")
		certBytes, keyBytes, err := cert.GenCertPair(
			defaultSelfSignedOrganization, cfg.TLSExtraIPs,
			cfg.TLSExtraDomains, cfg.TLSDisableAutofill,
			cfg.TLSCertDuration,
		)
		if err != nil {
			return tls.Certificate{}, nil, err
		}

		// Now that we have the certificate and key, we'll store them
		// to the file system.
		err = cert.WriteCertPair(
			cfg.TLSCertPath, cfg.TLSKeyPath, certBytes, keyBytes,
		)
		if err != nil {
			return tls.Certificate{}, nil, err
		}

		log.Infof("Done generating TLS certificates")
	}

	return cert.LoadCert(cfg.TLSCertPath, cfg.TLSKeyPath)
}
