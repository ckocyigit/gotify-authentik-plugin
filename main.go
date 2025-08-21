package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gotify/plugin-api"
)

const routeSuffix = "authentik"

func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		Name:        "Authentik Plugin",
		Description: "Plugin that enables Gotify to receive and understand the webhook structure from Authentik",
		ModulePath:  "github.com/ckocyigit/gotify-authentik-plugin",
		Author:      "Can Kocyigit <ckocyigit@ck98.de>",
		Website:     "https://cv.ck98.de",
	}
}

type Config struct {
	FriendlyName string `json:"friendly_name" yaml:"friendly_name"`
}

type Storage struct {
	Config Config `json:"config"`
}

type Plugin struct {
	userCtx        plugin.UserContext
	msgHandler     plugin.MessageHandler
	storageHandler plugin.StorageHandler
	config         *Config
	basePath       string
}

func (p *Plugin) DefaultConfig() interface{} {
	return &Config{
		FriendlyName: "",
	}
}

func (p *Plugin) ValidateAndSetConfig(conf interface{}) error {
	config, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type")
	}
	p.config = config
	return nil
}

func (p *Plugin) SetStorageHandler(h plugin.StorageHandler) {
	p.storageHandler = h
}

func (p *Plugin) Enable() error {
	storage := new(Storage)
	storageBytes, err := p.storageHandler.Load()
	if err != nil {
		return err
	}

	if len(storageBytes) > 0 {
		if err := json.Unmarshal(storageBytes, storage); err != nil {
			return err
		}
		p.config = &storage.Config
	}
	return nil
}

func (p *Plugin) Disable() error {
	return nil
}

func (p *Plugin) GetDisplay(location *url.URL) string {
	baseHost := ""
	if location != nil {
		baseHost = fmt.Sprintf("%s://%s", location.Scheme, location.Host)
	}
	webhookURL := baseHost + p.basePath + routeSuffix
	return fmt.Sprintf(`Steps to Configure Authentik Webhooks with Gotify:

	Create a Notification Transport in Authentik with the mode 'Webhook (generic)'.
	
	Copy this URL: %s and paste it in 'Webhook URL'.
	
	Keep the 'Webhook Mapping' field empty.
	
	Make sure to enable the 'Send once' option.
	
	Create a Notification Rule:
	- Assign the rule to a group, such as 'authentik Admins'.
	- Set the newly created transport as the delivery method.
	- Select Severity: 'Notice'.
	
	Create and bind two policies:
	- Policy 1: 
	  - Action: Login Failed
	  - The rest stays empty
	
	- Policy 2:
	  - Action: Login
	  - The rest stays empty
	
	Other event types are not currently supported for parsing but will still be displayed in Gotify, though without proper parsing.`, webhookURL)

}

func (p *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	p.msgHandler = h
}

func (p *Plugin) RegisterWebhook(basePath string, mux *gin.RouterGroup) {
	p.basePath = basePath
	mux.POST("/"+routeSuffix, p.webhookHandler)
}

func (p *Plugin) getMarkdownMsg(title string, message string, priority int, host string) plugin.Message {
	var instanceInfo string
	if p.config != nil && p.config.FriendlyName != "" {
		instanceInfo = fmt.Sprintf("Authentik instance: %s", p.config.FriendlyName)
	} else {
		instanceInfo = fmt.Sprintf("Authentik instance at: %s", host)
	}

	formattedMessage := fmt.Sprintf("%s\n\n```\n%s\n```", instanceInfo, message)

	return plugin.Message{
		Title:    title,
		Message:  formattedMessage,
		Priority: priority,
		Extras: map[string]interface{}{
			"client::display": map[string]interface{}{
				"contentType": "text/markdown",
			},
		},
	}
}

func (p *Plugin) webhookHandler(c *gin.Context) {
	var payload AuthentikWebhookPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		p.msgHandler.SendMessage(p.getMarkdownMsg(
			"Error parsing JSON message",
			err.Error(),
			7,
			c.Request.RemoteAddr,
		))
		return
	}

	title, message, priority := ReturnGotifyMessageFromAuthentikPayload(payload)

	p.msgHandler.SendMessage(p.getMarkdownMsg(
		title,
		message,
		priority,
		c.Request.RemoteAddr,
	))

}

func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{
		userCtx: ctx,
		config:  &Config{},
	}
}

func main() {
	panic("this should be built as go plugin")
}
