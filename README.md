# Gotify Authentik Plugin

This plugin enables [Gotify](https://gotify.net) to receive and process webhooks from [Authentik](https://goauthentik.io). It parses and formats the `login` and `login_failed` events into notifications for administrators, while other events are displayed in their raw form.

I just couldn’t get the mappings in Authentik to work properly with Gotify...

## Features
- **Login Events**: Get notified when users successfully log in.
- **Login Failed Events**: Receive detailed notifications when login attempts fail.
- **Custom Instance Name**: Configure a friendly name for your Authentik instance instead of showing the server address.

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ckocyigit/gotify-authentik-plugin.git
   ```

2. **Build the plugin:**
   Navigate to the project directory and build the Go plugin using:
   ```bash
   docker run --rm -v "$PWD/.:/proj" -w /proj gotify/build:1.22.4-linux-amd64             go build -a -installsuffix cgo -ldflags "-w -s" -buildmode=plugin -o plugin/authentik-plugin-amd64.so /proj
   ```

   Alternatively, you can download the prebuilt plugin from the [releases](https://github.com/ckocyigit/gotify-authentik-plugin/releases) page.

3. **Install the plugin into Gotify:**
   - Copy the generated `authentik-plugin-amd64.so` (or the downloaded release file) into the `plugins` folder of your Gotify instance.
   - (Optional) Set a friendly name for your Authentik instance in the plugin settings via the Gotify web interface.
   - The friendly name will replace the server address in notifications if configured.

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
   - Create two policies with the following actions (policy names can be freely chosen):
     - Policy 1: Action → `Login Failed`
     - Policy 2: Action → `Login`
   - The rest of the configuration can remain empty.

Other rules/policies are currently not supported natively, but they will still be displayed in Gotify without being parsed.

## Example `login_failed` Event

![Example login_failed Event](example.png)

## License
This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
