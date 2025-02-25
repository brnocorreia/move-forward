package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const (
	configFileName = ".move-forward.json"
	configRepo     = "https://raw.githubusercontent.com/move-forward/services/main"
)

type ServiceConfig struct {
	Name             string `json:"name"`
	ApiBaseURL       string `json:"apiBaseUrl"`
	WebSocketBaseURL string `json:"webSocketBaseUrl"`
	Description      string `json:"description"`
	PollIntervalSecs int    `json:"pollIntervalSeconds,omitempty"`
	MaxRetries       int    `json:"maxRetries,omitempty"`
}

type Config struct {
	Token          string                   `json:"token"`
	ExpireAt       time.Time                `json:"expireAt"`
	DeviceID       string                   `json:"deviceId"`
	ForwardURL     string                   `json:"forwardUrl"`
	CurrentService string                   `json:"currentService"`
	Services       map[string]ServiceConfig `json:"services"`
}

func ReadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{Services: make(map[string]ServiceConfig)}, nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Services == nil {
		config.Services = make(map[string]ServiceConfig)
	}

	return &config, nil
}

func WriteConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, configFileName), nil
}

func FetchServiceConfig(serviceName string) (*ServiceConfig, error) {
	url := fmt.Sprintf("%s/%s.json", configRepo, serviceName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("service %s not found (status %d)", serviceName, resp.StatusCode)
	}

	var service ServiceConfig
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, err
	}

	return &service, nil
}

func ListAvailableServices() ([]string, error) {
	url := fmt.Sprintf("%s/index.json", configRepo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch services list (status %d)", resp.StatusCode)
	}

	var services []string
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}

	return services, nil
}

func GetCurrentService(config *Config) (*ServiceConfig, error) {
	if config.CurrentService == "" {
		return nil, fmt.Errorf("no service configured")
	}

	service, exists := config.Services[config.CurrentService]
	if !exists {
		return nil, fmt.Errorf("service %s not found in local configuration", config.CurrentService)
	}

	return &service, nil
}

func PollForToken(deviceCode string, apiURL string, pollIntervalSecs int, maxRetries int) (string, error) {
	// Use default values if not specified
	if pollIntervalSecs <= 0 {
		pollIntervalSecs = 2 // Default to 2 seconds
	}
	if maxRetries <= 0 {
		maxRetries = 30 // Default to 30 retries
	}

	pollInterval := time.Duration(pollIntervalSecs) * time.Second

	for retries := 0; retries < maxRetries; retries++ {
		time.Sleep(pollInterval)

		body, _ := json.Marshal(map[string]string{"deviceCode": deviceCode})
		resp, err := http.Post(apiURL+"/token", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var result map[string]string
			json.NewDecoder(resp.Body).Decode(&result)
			if token, ok := result["token"]; ok {
				return token, nil
			}
		}
	}

	return "", fmt.Errorf("authorization timed out. Please try again")
}

func Login(apiURL string, pollIntervalSecs int, maxRetries int) {
	host, _ := os.Hostname()

	body, _ := json.Marshal(map[string]string{"host": host})
	resp, err := http.Post(apiURL+"/device-login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("❌ Login failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("🔗 Open the following link in your browser to authenticate:")
	fmt.Printf("👉 %s\n", result["verificationUri"])

	fmt.Println("⌛ Waiting for authorization...")
	token, err := PollForToken(result["deviceCode"], apiURL, pollIntervalSecs, maxRetries)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	config, err := ReadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to read config: %v", err)
	}

	config.Token = token

	if err := WriteConfig(config); err != nil {
		log.Fatalf("❌ Failed to save login token: %v", err)
	}

	fmt.Println("✅ Logged in successfully.")
}

func Listen(forwardURL, token, wsURL string) {
	fmt.Println("🌐 Starting webhook listener...")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+token, nil)
	if err != nil {
		log.Fatalf("❌ Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Webhook listener started")
	fmt.Printf("🔄 Forwarding events to: %s\n", forwardURL)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("❌ Webhook listener closed unexpectedly: %v", err)
		}

		fmt.Println("\n🌟 Received Webhook:")
		fmt.Println(string(message))

		resp, err := http.Post(forwardURL, "application/json", bytes.NewBuffer(message))
		if err != nil {
			log.Printf("❌ Failed to forward webhook: %v", err)
			continue
		}
		resp.Body.Close()
	}
}

func main() {
	var forwardURL string

	rootCmd := &cobra.Command{Use: "move-forward"}

	setupCmd := &cobra.Command{
		Use:   "setup [service-name]",
		Short: "Configure the CLI for a specific webhook service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := args[0]

			fmt.Printf("⚙️ Setting up for %s service...\n", serviceName)

			service, err := FetchServiceConfig(serviceName)
			if err != nil {
				log.Fatalf("❌ Setup failed: %v", err)
			}

			config, err := ReadConfig()
			if err != nil {
				log.Fatalf("❌ Failed to read config: %v", err)
			}

			config.Services[serviceName] = *service
			config.CurrentService = serviceName

			if err := WriteConfig(config); err != nil {
				log.Fatalf("❌ Failed to save configuration: %v", err)
			}

			fmt.Printf("✅ Setup complete for %s: %s\n", service.Name, service.Description)
			fmt.Println("🔑 Run 'move-forward login' to authenticate")
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available webhook services",
		Run: func(cmd *cobra.Command, args []string) {
			services, err := ListAvailableServices()
			if err != nil {
				log.Fatalf("❌ Failed to get services: %v", err)
			}

			fmt.Println("🔍 Available webhook services:")
			for _, service := range services {
				fmt.Printf("- %s\n", service)
			}
			fmt.Println("\nRun 'move-forward setup <service-name>' to configure")
		},
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to the configured webhook service",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadConfig()
			if err != nil {
				log.Fatalf("❌ Failed to read config: %v", err)
			}

			service, err := GetCurrentService(config)
			if err != nil {
				log.Fatalf("❌ %v. Please run 'move-forward setup <service-name>' first.", err)
			}

			Login(service.ApiBaseURL, service.PollIntervalSecs, service.MaxRetries)
		},
	}

	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for webhooks and forward them to your local server",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ReadConfig()
			if err != nil {
				log.Fatalf("❌ Failed to read config: %v", err)
			}

			service, err := GetCurrentService(config)
			if err != nil {
				log.Fatalf("❌ %v. Please run 'move-forward setup <service-name>' first.", err)
			}

			if config.Token == "" {
				log.Fatal("❌ You are not logged in. Please run 'move-forward login' first.")
			}

			if forwardURL == "" && config.ForwardURL != "" {
				forwardURL = config.ForwardURL
			} else if forwardURL != "" {
				config.ForwardURL = forwardURL
				WriteConfig(config)
			}

			Listen(forwardURL, config.Token, service.WebSocketBaseURL)
		},
	}

	listenCmd.Flags().StringVarP(&forwardURL, "forward", "f", "", "Local server URL to forward webhooks")
	listenCmd.MarkFlagRequired("forward")

	rootCmd.AddCommand(setupCmd, listCmd, loginCmd, listenCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
