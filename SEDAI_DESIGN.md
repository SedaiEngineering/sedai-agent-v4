sedai-agent-v4: Developer README

This document outlines the architecture, functionality, and modifications of sedai-agent-v4, an internal fork of the open-source chisel project. The goal of this fork is to create a hardened, minimal, and special-purpose secure tunneling tool tailored specifically for the SEDAI SaaS connectivity model.

1. Vision & Core Architecture
sedai-agent-v4 is not a general-purpose tunneling tool. It has been stripped down to serve a single architectural pattern:

Server: A single sedai-agent-v4 server instance runs within a dedicated tenant namespace in the SEDAI SaaS Kubernetes environment. It acts as the secure ingress point for a specific customer.

Client: A single sedai-agent-v4 client instance runs as a containerized agent within the customer's private environment (VPC, on-prem).

Connection Flow: The client always initiates an outbound, reverse tunnel connection to the server. This allows the SEDAI SaaS platform to access specific, pre-approved services (e.g., databases, VMs) in the customer's environment without requiring the customer to open any inbound firewall ports.

The primary design principle is simplicity and reduction of attack surface. By removing all non-essential features, we create a more secure and easier-to-maintain tool.

2. Summary of Major Changes
A developer familiar with the original chisel should be aware of the following key modifications:

Reverse-Only Functionality: The codebase has been stripped of all logic related to forward tunnels and SOCKS server modes. This fork only supports reverse tunnels. The --reverse flag is now the default and only behavior.

Simplified Server Authentication: The --auth and --authfile flags have been removed from the server. Authentication is now handled exclusively by passing a JSON string containing user credentials via a command-line flag (--auth-json) or an environment variable (CHISEL_AUTH_JSON).

Removed --backend Option: The --backend flag, which allowed for an experimental SSH-based backend, has been completely removed to simplify the codebase and eliminate an unnecessary dependency.

Streamlined Codebase: All code paths, configuration options, and logic not directly related to the reverse tunnel use case have been excised.

3. Detailed Changes & Developer Guide
3.1. Reverse-Only Mode
The concept of tunnel direction is no longer configurable. The tool is hardcoded to operate in a reverse tunnel mode.

Impact on Server: The server is permanently in --reverse mode. It will listen for client connections and expect them to define the remote-to-local port mappings.

Impact on Client: The client no longer needs to specify a direction. The R prefix for remote forwards is the only accepted syntax.

Developer Notes:

Review main.go and the cmd/ directory to see the removal of command parsing for other modes.

The core connection handling logic in the server and client packages has been simplified to remove conditional checks for tunnel direction.

3.2. New Authentication Mechanism
To better align with cloud-native deployment patterns (e.g., passing secrets from Kubernetes), file-based authentication has been replaced.

The Old Way (Removed):

# --authuser user:pass
# --authfile /path/to/users.json

The New Way (Mandatory):

Authentication credentials are provided as a JSON string.

JSON Format:

{
  "users": [
    {
      "name": "customer-user",
      "password": "a-very-strong-secret-password"
    }
  ]
}

Usage:

1. Via Environment Variable (Recommended for Kubernetes):

export CHISEL_AUTH_JSON='{"users":[{"name":"customer-user","password":"a-very-strong-secret-password"}]}'
sedai-agent-v4 server --port 443

2. Via Command-Line Flag:

sedai-agent-v4 server --port 443 --auth-json '{"users":[{"name":"customer-user","password":"a-very-strong-secret-password"}]}'

Developer Notes:

See config.go for the removal of the AuthFile and Auth fields and the addition of the new AuthJSON field.

The server's initialization logic has been modified to parse this JSON string instead of reading from a file.

3.3. Removed Command-Line Flags
To reduce complexity, the following flags have been completely removed and are no longer available:

--backend: The alternative SSH backend is not needed.

--socks5: The SOCKS5 server functionality has been removed.

--reverse: This is now the default behavior and the flag is no longer necessary.

Any flags related to forward tunnels (e.g., those that don't use the R: prefix).

4. Usage Examples
Server (SEDAI SaaS Namespace)
Run the server, providing the auth JSON via an environment variable. Note the use of sedai-agent-v4 server.

# Set the auth secret in the environment
export CHISEL_AUTH_JSON='{"users":[{"name":"customer01","password":"super-secret-token-for-customer01"}]}'

# Run the server
./sedai-agent-v4 server --port 443 --hostname your-tenant.saas.com

Client (Customer Environment Container)
The client connects to the server and defines the reverse tunnel, mapping the server's remote port 9000 to the customer's internal database on port 5432. Note the use of sedai-agent-v4 client.

# Run the client agent
./sedai-agent-v4 client \
  --user "customer01" \
  --password "super-secret-token-for-customer01" \
  your-tenant.saas.com:443 \
  R:9000:db1.customer.internal:5432

The api_container in the SEDAI namespace can now connect to localhost:9000 to access the customer's database.
