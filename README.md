# Gotify Authentik Plugin

This plugin enables [Gotify](https://gotify.net) to receive and process webhooks from [Authentik](https://goauthentik.io). It parses and formats the `login`, `login_failed`, and `logout` events into notifications for administrators, while other events are displayed in their raw form.

I just couldn’t get the mappings in Authentik to work properly with Gotify...

## Features
- **Login Events**: Get notified when users successfully log in, including location, coordinates, network and client details.
- **Login Failed Events**: Receive detailed notifications when login attempts fail, including stage, network and client details.
- **Logout Events**: Receive logout notifications, including logout reason and binding details when Authentik provides them.
- **Custom Instance Name**: Configure a friendly name for your Authentik instance instead of showing the server address.

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ckocyigit/gotify-authentik-plugin.git
   ```

2. **Build the plugin:**
   Navigate to the project directory and build the Go plugin using:
   ```bash
   docker run --rm -v "$PWD/.:/proj" -w /proj gotify/build:1.26.0-linux-amd64 \
     go build -a -installsuffix cgo -ldflags "-w -s" -buildmode=plugin -o plugin/authentik-plugin-amd64.so /proj
   ```

   The build script reads the required Go toolchain from `go.mod` and selects the matching `gotify/build` image automatically.

   Alternatively, you can download the prebuilt plugin from the [releases](https://github.com/ckocyigit/gotify-authentik-plugin/releases) page.

3. **Install the plugin into Gotify:**
   - Copy the generated `authentik-plugin-amd64.so` (or the downloaded release file) into the `plugins` folder of your Gotify instance.
   - (Optional) Set a friendly name for your Authentik instance in the plugin settings via the Gotify web interface.
   - The friendly name will replace the server address in notifications if configured.
   - The displayed client IP prefers `client_ip` from the Authentik payload, then falls back to the payload's network range or the webhook source address when no dedicated client IP is present.

## Configuration in Authentik

To configure the webhook transport in Authentik, follow these steps:

1. **Create a Notification Transport in Authentik with mode `Webhook (generic)`.**

2. **Copy the webhook URL from Gotify:**
   - Copy the webhook URL from the Gotify plugin settings page.

3. **Webhook Mapping:**
   - Leave the Webhook Mapping empty.

4. **Enable the `Send once` option.**

5. **Create a Notification Rule:**
   - Example: Create a rule for the *authentik Admins* group and enable the newly created transport.

6. **Set Severity Level:**
   - Select severity `Notice`.

7. **Create and bind Policies:**
   - Create three policies with the following actions (policy names can be freely chosen):
   - Policy 1: Action → `Login Failed`
   - Policy 2: Action → `Login`
   - Policy 3: Action → `Logout`
   - The rest of the configuration can remain empty.

Other rules/policies are currently not supported natively, but they will still be displayed in Gotify without being parsed.

## Maintenance

To sync this plugin with the latest Gotify server release locally, run:

```bash
bash ./tidy.sh
```

To sync against a specific Gotify release instead, pass the tag explicitly:

```bash
bash ./tidy.sh v2.9.1
```

The sync script downloads the matching Gotify server `go.mod`, caps shared dependency versions with `gomod-cap`, aligns the plugin module's `go` and `toolchain` directives, and runs `go mod tidy`.

The repository also includes an automated sync workflow in `.github/workflows/sync-gotify-release.yml`. For direct commits to `main`, the workflow's `GITHUB_TOKEN` must be allowed to write to the branch by your branch protection or ruleset.

## Example `login_failed` Event

![Example login_failed Event](example.png)

## License
This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
