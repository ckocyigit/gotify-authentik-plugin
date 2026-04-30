package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AuthMethodArgs struct {
	MFADevices   []MFADevice `json:"mfa_devices"`
	KnownDevice  string      `json:"known_device,omitempty"`
}
type AuthentikWebhookPayload struct {
	Body              string `json:"body"`
	EventUserEmail    string `json:"event_user_email"`
	EventUserUsername string `json:"event_user_username"`
	Severity          string `json:"severity"`
	UserEmail         string `json:"user_email"`
	UserUsername      string `json:"user_username"`
}

type LoginFailedData struct {
	ASN         ASN         `json:"asn"`
	ClientIP    string      `json:"client_ip,omitempty"`
	Geo         Geo         `json:"geo"`
	Stage       Stage       `json:"stage"`
	Username    string      `json:"username"`
	DeviceClass string      `json:"device_class,omitempty"`
	Password    string      `json:"password,omitempty"`
	HTTPRequest HTTPRequest `json:"http_request"`
}

type LoginData struct {
	ASN            ASN            `json:"asn"`
	ClientIP       string         `json:"client_ip,omitempty"`
	Geo            Geo            `json:"geo"`
	AuthMethod     string         `json:"auth_method"`
	HTTPRequest    HTTPRequest    `json:"http_request"`
	AuthMethodArgs AuthMethodArgs `json:"auth_method_args"`
}

type LogoutData struct {
	ASN          ASN           `json:"asn"`
	Binding      LogoutBinding `json:"binding,omitempty"`
	ClientIP     string        `json:"client_ip,omitempty"`
	Geo          Geo           `json:"geo"`
	HTTPRequest  HTTPRequest   `json:"http_request"`
	IP           IPChange      `json:"ip,omitempty"`
	LogoutReason string        `json:"logout_reason,omitempty"`
}

type ASN struct {
	ASN     int    `json:"asn"`
	ASOrg   string `json:"as_org"`
	Network string `json:"network"`
}

type Geo struct {
	Lat       float64 `json:"lat"`
	City      string  `json:"city"`
	Long      float64 `json:"long"`
	Country   string  `json:"country"`
	Continent string  `json:"continent"`
}

type Stage struct {
	PK        string `json:"pk"`
	App       string `json:"app"`
	Name      string `json:"name"`
	ModelName string `json:"model_name"`
}

type LogoutBinding struct {
	Reason string `json:"reason,omitempty"`
}

type IPChange struct {
	New      string `json:"new,omitempty"`
	Previous string `json:"previous,omitempty"`
}

type HTTPRequest struct {
	Args      map[string]string `json:"args"`
	Path      string            `json:"path"`
	Method    string            `json:"method"`
	RequestID string            `json:"request_id"`
	UserAgent string            `json:"user_agent"`
}

type MFADevice struct {
	PK        int    `json:"pk"`
	App       string `json:"app"`
	Name      string `json:"name"`
	ModelName string `json:"model_name"`
}

func ReturnGotifyMessageFromAuthentikPayload(payload AuthentikWebhookPayload, sourceIP string) (string, string, int) {
	if strings.HasPrefix(payload.Body, "login_failed: ") {
		var data LoginFailedData
		bodyContent := strings.TrimPrefix(payload.Body, "login_failed: ")
		bodyContent = strings.ReplaceAll(bodyContent, "'", "\"")

		if err := json.Unmarshal([]byte(bodyContent), &data); err != nil {
			return "Error parsing login_failed data", err.Error(), 7
		}

		title := fmt.Sprintf("Login failed for %s", data.Username)
		message := formatLoginFailedMessage(data, sourceIP)

		return title, message, 8

	} else if strings.HasPrefix(payload.Body, "login: ") {
		var data LoginData
		bodyContent := strings.TrimPrefix(payload.Body, "login: ")
		bodyContent = strings.ReplaceAll(bodyContent, "'", "\"")
		bodyContent = strings.ReplaceAll(bodyContent, "True", `"true"`)
		bodyContent = strings.ReplaceAll(bodyContent, "False", `"false"`)

		if err := json.Unmarshal([]byte(bodyContent), &data); err != nil {
			return "Error parsing login data", err.Error(), 7
		}

		username := preferredUsername(payload)
		title := fmt.Sprintf("%s just logged in", username)
		message := formatLoginMessage(username, data, sourceIP)

		return title, message, 5

	} else if strings.HasPrefix(payload.Body, "logout: ") {
		var data LogoutData
		bodyContent := strings.TrimPrefix(payload.Body, "logout: ")
		bodyContent = strings.ReplaceAll(bodyContent, "'", "\"")
		bodyContent = strings.ReplaceAll(bodyContent, "True", `"true"`)
		bodyContent = strings.ReplaceAll(bodyContent, "False", `"false"`)

		if err := json.Unmarshal([]byte(bodyContent), &data); err != nil {
			return "Error parsing logout data", err.Error(), 7
		}

		username := preferredUsername(payload)
		title := fmt.Sprintf("%s logged out", username)
		message := formatLogoutMessage(username, data, sourceIP)

		return title, message, 5

	} else {
		title := "Unrecognized Event"
		message := payload.Body
		return title, message, 5
	}
}

func preferredUsername(payload AuthentikWebhookPayload) string {
	if payload.EventUserUsername != "" {
		return payload.EventUserUsername
	}
	if payload.UserUsername != "" {
		return payload.UserUsername
	}

	return "Unknown user"
}

func formatLoginFailedMessage(data LoginFailedData, sourceIP string) string {
	clientAddress := formatClientAddress(data.ClientIP, data.ASN.Network, sourceIP)
	lines := []string{
		fmt.Sprintf("Login attempt failed for user: %s", data.Username),
	}

	lines = appendDetails(lines,
		"Location", formatLocation(data.Geo),
		"Coordinates", formatCoordinates(data.Geo),
		"Client IP", clientAddress,
		"Network", formatNetwork(data.ASN, clientAddress),
		"Client", describeUserAgent(data.HTTPRequest.UserAgent),
		"Device Class", data.DeviceClass,
		"Stage", data.Stage.Name,
		"Stage Model", data.Stage.ModelName,
		"RequestID", data.HTTPRequest.RequestID,
	)

	return strings.Join(lines, "\n")
}

func formatLogoutMessage(username string, data LogoutData, sourceIP string) string {
	clientAddress := formatLogoutClientAddress(data, sourceIP)
	lines := []string{
		fmt.Sprintf("Successful logout for user: %s", username),
	}

	lines = appendDetails(lines,
		"Client IP", clientAddress,
		"Location", formatLocation(data.Geo),
		"Coordinates", formatCoordinates(data.Geo),
		"Network", formatNetwork(data.ASN, clientAddress),
		"Previous IP", formatPreviousIP(data.IP, clientAddress),
		"Client", describeUserAgent(data.HTTPRequest.UserAgent),
		"Request", formatRequest(data.HTTPRequest),
		"Logout Reason", data.LogoutReason,
		"Binding Reason", data.Binding.Reason,
		"RequestID", data.HTTPRequest.RequestID,
	)

	return strings.Join(lines, "\n")
}

func formatLoginMessage(username string, data LoginData, sourceIP string) string {
	clientAddress := formatClientAddress(data.ClientIP, data.ASN.Network, sourceIP)
	lines := []string{
		fmt.Sprintf("Successful login for user: %s", username),
	}

	lines = appendDetails(lines,
		"Location", formatLocation(data.Geo),
		"Coordinates", formatCoordinates(data.Geo),
		"Client IP", clientAddress,
		"Network", formatNetwork(data.ASN, clientAddress),
		"Auth Method", data.AuthMethod,
		"Client", describeUserAgent(data.HTTPRequest.UserAgent),
		"Known Device", data.AuthMethodArgs.KnownDevice,
		"MFA Devices", formatMFADevices(data.AuthMethodArgs.MFADevices),
		"RequestID", data.HTTPRequest.RequestID,
	)

	return strings.Join(lines, "\n")
}

func appendDetails(lines []string, values ...string) []string {
	for index := 0; index+1 < len(values); index += 2 {
		label := strings.TrimSpace(values[index])
		value := strings.TrimSpace(values[index+1])
		if label == "" || value == "" {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s: %s", label, value))
	}

	return lines
}

func formatLocation(geo Geo) string {
	parts := []string{}

	if geo.City != "" {
		parts = append(parts, geo.City)
	}
	if geo.Country != "" {
		parts = append(parts, geo.Country)
	}
	if geo.Continent != "" {
		parts = append(parts, geo.Continent)
	}

	return strings.Join(parts, ", ")
}

func formatCoordinates(geo Geo) string {
	if geo.Lat == 0 && geo.Long == 0 {
		return ""
	}

	return fmt.Sprintf("%.4f, %.4f", geo.Lat, geo.Long)
}

func formatNetwork(asn ASN, clientAddress string) string {
	parts := []string{}

	if asn.ASOrg != "" {
		parts = append(parts, asn.ASOrg)
	}
	if asn.Network != "" && asn.Network != clientAddress {
		parts = append(parts, asn.Network)
	}
	if asn.ASN != 0 {
		parts = append(parts, fmt.Sprintf("AS%d", asn.ASN))
	}

	return strings.Join(parts, " | ")
}

func formatClientAddress(clientIP string, network string, sourceIP string) string {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP != "" {
		return clientIP
	}

	network = strings.TrimSpace(network)
	if network != "" {
		return network
	}

	return strings.TrimSpace(sourceIP)
}

func formatLogoutClientAddress(data LogoutData, sourceIP string) string {
	if data.ClientIP != "" {
		return strings.TrimSpace(data.ClientIP)
	}
	if data.IP.New != "" {
		return strings.TrimSpace(data.IP.New)
	}

	return formatClientAddress("", data.ASN.Network, sourceIP)
}

func formatPreviousIP(ipChange IPChange, clientAddress string) string {
	previous := strings.TrimSpace(ipChange.Previous)
	if previous == "" || previous == clientAddress {
		return ""
	}

	return previous
}

func formatRequest(httpRequest HTTPRequest) string {
	method := strings.TrimSpace(httpRequest.Method)
	path := strings.TrimSpace(httpRequest.Path)

	switch {
	case method != "" && path != "":
		return fmt.Sprintf("%s %s", method, path)
	case method != "":
		return method
	default:
		return path
	}
}

func formatMFADevices(devices []MFADevice) string {
	if len(devices) == 0 {
		return ""
	}

	formattedDevices := make([]string, 0, len(devices))
	for _, device := range devices {
		name := strings.TrimSpace(device.Name)
		model := strings.TrimSpace(device.ModelName)

		switch {
		case name != "" && model != "" && name != model:
			formattedDevices = append(formattedDevices, fmt.Sprintf("%s (%s)", name, model))
		case name != "":
			formattedDevices = append(formattedDevices, name)
		case model != "":
			formattedDevices = append(formattedDevices, model)
		}
	}

	return strings.Join(formattedDevices, ", ")
}

func describeUserAgent(userAgent string) string {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return ""
	}

	browser := detectBrowser(userAgent)
	osName := detectOperatingSystem(userAgent)

	switch {
	case browser != "" && osName != "":
		return fmt.Sprintf("%s on %s", browser, osName)
	case browser != "":
		return browser
	case osName != "":
		return osName
	default:
		return userAgent
	}
}

func detectBrowser(userAgent string) string {
	browserTokens := []struct {
		name  string
		token string
	}{
		{name: "Edge", token: "Edg/"},
		{name: "Firefox", token: "Firefox/"},
		{name: "Chrome", token: "Chrome/"},
	}

	for _, browser := range browserTokens {
		if version := majorVersion(userAgent, browser.token); version != "" {
			return fmt.Sprintf("%s %s", browser.name, version)
		}
	}

	if strings.Contains(userAgent, "Safari/") {
		if version := majorVersion(userAgent, "Version/"); version != "" {
			return fmt.Sprintf("Safari %s", version)
		}
		return "Safari"
	}

	return ""
}

func detectOperatingSystem(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Windows NT 10.0"):
		return "Windows 10/11"
	case strings.Contains(userAgent, "Windows NT 6.3"):
		return "Windows 8.1"
	case strings.Contains(userAgent, "Windows"):
		return "Windows"
	case strings.Contains(userAgent, "Android"):
		return "Android"
	case strings.Contains(userAgent, "iPhone"):
		return "iPhone"
	case strings.Contains(userAgent, "iPad"):
		return "iPad"
	case strings.Contains(userAgent, "Mac OS X"):
		return "macOS"
	case strings.Contains(userAgent, "Linux"):
		return "Linux"
	default:
		return ""
	}
}

func majorVersion(userAgent string, token string) string {
	start := strings.Index(userAgent, token)
	if start == -1 {
		return ""
	}

	version := userAgent[start+len(token):]
	separator := strings.IndexAny(version, " ;)")
	if separator != -1 {
		version = version[:separator]
	}

	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}

	parts := strings.SplitN(version, ".", 2)
	return parts[0]
}
