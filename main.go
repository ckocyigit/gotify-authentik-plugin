package main

import (
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

type Plugin struct {
	userCtx    plugin.UserContext
	msgHandler plugin.MessageHandler
	basePath   string
}

func (p *Plugin) Enable() error {
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

func getMarkdownMsg(title string, message string, priority int, host string) plugin.Message {
	formattedMessage := fmt.Sprintf("Authentik instance at: %s\n\n```\n%s\n```", host, message)

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
		p.msgHandler.SendMessage(getMarkdownMsg(
			"Error parsing JSON message",
			err.Error(),
			7,
			c.Request.RemoteAddr,
		))
		return
	}

	title, message, priority := ReturnGotifyMessageFromAuthentikPayload(payload)

	p.msgHandler.SendMessage(getMarkdownMsg(
		title,
		message,
		priority,
		c.Request.RemoteAddr,
	))

}

func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{
		userCtx: ctx,
	}
}

func main() {
	panic("this should be built as go plugin")
}
