package main

import (
	"strings"
	"testing"
)

func TestReturnGotifyMessageFromAuthentikPayloadParsesLoginBodyWithMixedQuotes(t *testing.T) {
	payload := AuthentikWebhookPayload{
		Body:              `login: {'asn': {'asn': 64512, 'as_org': 'CK98 ISP', 'network': '203.0.113.0/24'}, 'client_ip': '203.0.113.42', 'geo': {'lat': 48.1351, 'city': 'Munich', 'long': 11.582, 'country': 'Germany', 'continent': 'Europe'}, 'auth_method': 'password', 'http_request': {'args': {}, 'path': '/if/flow/default-authentication-flow/', 'method': 'POST', 'request_id': 'req-123', 'user_agent': "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 Version/17.4 Mobile/15E148 Safari/604.1"}, 'auth_method_args': {'known_device': True, 'mfa_devices': [{'pk': 1, 'app': 'authentik_stages_authenticator_totp', 'name': "Chris's iPhone", 'model_name': 'TOTP Device'}]}}`,
		EventUserUsername: "ck",
	}

	title, message, priority := ReturnGotifyMessageFromAuthentikPayload(payload, "198.51.100.10")

	if title != "ck just logged in" {
		t.Fatalf("unexpected title: %q", title)
	}

	if priority != 5 {
		t.Fatalf("unexpected priority: %d", priority)
	}

	expectedFragments := []string{
		"Successful login for user: ck",
		"Client IP: 203.0.113.42",
		"Auth Method: password",
		"Known Device: true",
		"MFA Devices: Chris's iPhone (TOTP Device)",
		"Client: Safari 17 on iPhone",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(message, fragment) {
			t.Fatalf("expected %q in message %q", fragment, message)
		}
	}
}
