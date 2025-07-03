package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"opnsense-auto-dns/internal/api/opnsense"
	"opnsense-auto-dns/internal/logger"
)

type Config struct {
	OPNsenseHost      string   `json:"opnsense_host"`
	OPNsenseAPIKey    string   `json:"opnsense_api_key"`
	OPNsenseAPISecret string   `json:"opnsense_api_secret"`
	Domain            string   `json:"domain"`
	Hostnames         []string `json:"hostnames,omitempty"`
	IPAddress         string   `json:"ip_address,omitempty"`
}

var (
	configFile string
	interval   int
	loop       bool
	ignoreCert bool

	opnsenseHost      string
	opnsenseAPIKey    string
	opnsenseAPISecret string
	domain            string
	ipAddress         string
	hostnames         []string
)

var autoUpdaterCmd = &cobra.Command{
	Use:   "auto-updater",
	Short: "Automatically update DNS records in OPNsense",
	Long: `Auto-updater command that updates DNS records in OPNsense unbound.
It can run once or in a continuous loop to keep DNS records updated.

This command automatically detects your current IP address and updates DNS records
in OPNsense to keep your hostnames pointing to the correct IP address.

Configuration precedence (highest to lowest):
1. Environment variables
2. Command line flags
3. Config file (--config flag)

IP address can be specified via config file (ip_address), environment variable (IP_ADDRESS),
or command line flag (--ip-address). If not provided, the current machine's IP will be detected automatically.

Environment variables:
- OPNSENSE_HOST, OPNSENSE_API_KEY, OPNSENSE_API_SECRET
- HOSTNAMES (comma-separated list), DOMAIN, IP_ADDRESS
- INTERVAL, LOOP, IGNORE_CERT

A config file can be specified using the --config flag, or configuration can be provided 
via environment variables and command line flags.

Hostnames can be specified as an array in config file or via --hostnames flag.
If no hostnames are specified, the machine hostname will be used.

Examples:
  # Run once with config file
  opnsense-auto-dns auto-updater --config config.json

  # Run in continuous loop
  opnsense-auto-dns auto-updater --config config.json --loop --interval 10

  # Use environment variables only
  opnsense-auto-dns auto-updater`,
	Run: runAutoUpdater,
}

func init() {
	rootCmd.AddCommand(autoUpdaterCmd)

	autoUpdaterCmd.Flags().StringVar(&configFile, "config", "", "config file path")
	autoUpdaterCmd.Flags().IntVar(&interval, "interval", 5, "update interval in minutes (when using --loop)")
	autoUpdaterCmd.Flags().BoolVar(&loop, "loop", false, "run in continuous loop")
	autoUpdaterCmd.Flags().BoolVar(&ignoreCert, "ignore-cert", false, "ignore certificate validation")

	autoUpdaterCmd.Flags().StringVar(&opnsenseHost, "opnsense-host", "", "OPNsense host (overrides config file)")
	autoUpdaterCmd.Flags().StringVar(&opnsenseAPIKey, "opnsense-api-key", "", "OPNsense API key (overrides config file)")
	autoUpdaterCmd.Flags().StringVar(&opnsenseAPISecret, "opnsense-api-secret", "", "OPNsense API secret (overrides config file)")
	autoUpdaterCmd.Flags().StringVar(&domain, "domain", "", "domain (overrides config file)")
	autoUpdaterCmd.Flags().StringVar(&ipAddress, "ip-address", "", "IP address (overrides config file)")
	autoUpdaterCmd.Flags().StringSliceVar(&hostnames, "hostnames", []string{}, "hostnames (overrides config file)")
}

func runAutoUpdater(cmd *cobra.Command, args []string) {
	if envInterval := os.Getenv("INTERVAL"); envInterval != "" {
		if parsedInterval, err := strconv.Atoi(envInterval); err == nil {
			interval = parsedInterval
			logger.Debug("Overriding interval from environment", "value", interval)
		} else {
			logger.Warn("Invalid INTERVAL environment variable", "value", envInterval, "err", err)
		}
	}

	if envLoop := os.Getenv("LOOP"); envLoop != "" {
		if parsedLoop, err := strconv.ParseBool(envLoop); err == nil {
			loop = parsedLoop
			logger.Debug("Overriding loop from environment", "value", loop)
		} else {
			logger.Warn("Invalid LOOP environment variable", "value", envLoop, "err", err)
		}
	}

	if envIgnoreCert := os.Getenv("IGNORE_CERT"); envIgnoreCert != "" {
		if parsedIgnoreCert, err := strconv.ParseBool(envIgnoreCert); err == nil {
			ignoreCert = parsedIgnoreCert
			logger.Debug("Overriding ignore-cert from environment", "value", ignoreCert)
		} else {
			logger.Warn("Invalid IGNORE_CERT environment variable", "value", envIgnoreCert, "err", err)
		}
	}

	config := loadConfig()

	if loop {
		logger.Info("Starting auto-updater in loop mode", "interval", interval)
		for {
			updateDNS(config)
			time.Sleep(time.Duration(interval) * time.Minute)
		}
	} else {
		updateDNS(config)
	}
}

func loadConfig() *Config {
	var config Config

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			logger.Fatal("Error reading config file", "path", configFile, "err", err)
		}
		logger.Debug("Using config file", "path", configFile)

		if err := json.Unmarshal(data, &config); err != nil {
			logger.Fatal("Error parsing config", "err", err)
		}
	} else {
		logger.Debug("No config file provided, using environment variables and command line flags only")
	}

	if opnsenseHost != "" {
		config.OPNsenseHost = opnsenseHost
		logger.Debug("Overriding opnsense_host from command line", "value", opnsenseHost)
	}
	if opnsenseAPIKey != "" {
		config.OPNsenseAPIKey = opnsenseAPIKey
		logger.Debug("Overriding opnsense_api_key from command line")
	}
	if opnsenseAPISecret != "" {
		config.OPNsenseAPISecret = opnsenseAPISecret
		logger.Debug("Overriding opnsense_api_secret from command line")
	}
	if domain != "" {
		config.Domain = domain
		logger.Debug("Overriding domain from command line", "value", domain)
	}
	if ipAddress != "" {
		config.IPAddress = ipAddress
		logger.Debug("Overriding ip_address from command line", "value", ipAddress)
	}
	if len(hostnames) > 0 {
		config.Hostnames = hostnames
		logger.Debug("Overriding hostnames from command line", "hostnames", hostnames)
	}

	if envHost := os.Getenv("OPNSENSE_HOST"); envHost != "" {
		config.OPNsenseHost = envHost
		logger.Debug("Overriding opnsense_host from environment", "value", envHost)
	}
	if envAPIKey := os.Getenv("OPNSENSE_API_KEY"); envAPIKey != "" {
		config.OPNsenseAPIKey = envAPIKey
		logger.Debug("Overriding opnsense_api_key from environment")
	}
	if envAPISecret := os.Getenv("OPNSENSE_API_SECRET"); envAPISecret != "" {
		config.OPNsenseAPISecret = envAPISecret
		logger.Debug("Overriding opnsense_api_secret from environment")
	}
	if envDomain := os.Getenv("DOMAIN"); envDomain != "" {
		config.Domain = envDomain
		logger.Debug("Overriding domain from environment", "value", envDomain)
	}
	if envHostnames := os.Getenv("HOSTNAMES"); envHostnames != "" {
		config.Hostnames = strings.Split(envHostnames, ",")
		logger.Debug("Overriding hostnames from environment", "hostnames", config.Hostnames)
	}
	if envIPAddress := os.Getenv("IP_ADDRESS"); envIPAddress != "" {
		config.IPAddress = envIPAddress
		logger.Debug("Overriding ip_address from environment", "value", envIPAddress)
	}

	if config.OPNsenseHost == "" {
		logger.Fatal("opnsense_host is required")
	}
	if config.OPNsenseAPIKey == "" {
		logger.Fatal("opnsense_api_key is required")
	}
	if config.OPNsenseAPISecret == "" {
		logger.Fatal("opnsense_api_secret is required")
	}
	if config.Domain == "" {
		logger.Fatal("domain is required")
	}

	return &config
}

func getMachineHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get machine hostname: %v", err)
	}

	logger.Debug("Fetched machine hostname", "hostname", hostname)
	return hostname, nil
}

func getHostnamesToUse(config *Config) ([]string, error) {
	if len(config.Hostnames) > 0 {
		logger.Debug("Using configured hostnames", "hostnames", config.Hostnames)
		return config.Hostnames, nil
	}

	hostname, err := getMachineHostname()
	if err != nil {
		return nil, err
	}

	logger.Debug("Using machine hostname", "hostname", hostname)
	return []string{hostname}, nil
}

func getCurrentIP(config *Config) (string, error) {
	if config.IPAddress != "" {
		logger.Debug("Using provided IP address", "ip", config.IPAddress)
		return config.IPAddress, nil
	}

	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()

	logger.Debug("Fetched local IP", "ip", ip)
	return ip, nil
}

func updateDNS(config *Config) {
	currentIP, err := getCurrentIP(config)
	if err != nil {
		logger.Error("Error getting current IP", "err", err)
		return
	}

	hostnamesToUse, err := getHostnamesToUse(config)
	if err != nil {
		logger.Error("Error getting hostnames to use", "err", err)
		return
	}

	logger.Info("Updating DNS records", "hostnames", hostnamesToUse, "ip", currentIP)

	client := opnsense.NewClient(config.OPNsenseHost, config.OPNsenseAPIKey, config.OPNsenseAPISecret, ignoreCert)

	for _, hostname := range hostnamesToUse {
		if err := updateDNSForHostname(client, hostname, config.Domain, currentIP); err != nil {
			logger.Error("Error updating DNS for hostname", "hostname", hostname, "err", err)
		}
	}
}

func updateDNSForHostname(client *opnsense.Client, hostname, domain, currentIP string) error {
	existingRecord, err := client.Unbound.GetExistingDNSRecord(hostname, domain)
	if err != nil {
		return fmt.Errorf("error getting existing DNS record: %v", err)
	}

	var oldIP string
	if existingRecord != nil {
		oldIP = existingRecord.Server
		logger.Debug("Found existing DNS record", "hostname", hostname, "old_ip", oldIP, "uuid", existingRecord.UUID)
	} else {
		oldIP = "none"
		logger.Debug("No existing DNS record found, will create new one", "hostname", hostname)
	}

	if oldIP == currentIP {
		logger.Debug("IP unchanged, skipping update", "hostname", hostname, "ip", currentIP)
		return nil
	}

	logger.Info("IP changed, updating DNS", "hostname", hostname, "old_ip", oldIP, "new_ip", currentIP)

	if existingRecord != nil {
		if err := client.Unbound.UpdateDNSRecord(existingRecord, hostname, domain, currentIP); err != nil {
			return fmt.Errorf("error updating DNS record: %v", err)
		}
		logger.Info("Successfully updated DNS record", "hostname", hostname, "domain", domain, "ip", currentIP)
	} else {
		if err := client.Unbound.CreateDNSRecord(hostname, domain, currentIP); err != nil {
			return fmt.Errorf("error creating DNS record: %v", err)
		}
		logger.Info("Successfully created DNS record", "hostname", hostname, "domain", domain, "ip", currentIP)
	}

	return nil
}
