# sedai-agent-v4

This document provides instructions for compiling and running `sedai-agent-v4`. For design and architecture details, see `SEDAI_DESIGN.md`.

## Compilation

To compile the agent, run the following command from the root of the repository:

```bash
go build -o sedai-agent-v4 .
```

This will create a `sedai-agent-v4` executable in the current directory.

## Usage

`sedai-agent-v4` operates in two modes: `server` and `client`.

### Server

The server runs in the SEDAI SaaS environment and listens for incoming client connections.

**To run the server:**

You must provide user authentication credentials as a JSON string, either via an environment variable or a command-line flag.

**Option 1: Using environment variable (recommended)**
1.  Set `CHISEL_AUTH_JSON` environment variable:
    ```bash
    export CHISEL_AUTH_JSON='{"users":[{"name":"customer01","password":"super-secret-token-for-customer01"}]}'
    ```
2.  Run the server:
    ```bash
    ./sedai-agent-v4 server --port 443
    ```

**Option 2: Using `--auth-json` flag**
Run the server executable, providing the JSON string.
    ```bash
    ./sedai-agent-v4 server --port 443 --auth-json '{"users":[{"name":"customer01","password":"super-secret-token-for-customer01"}]}'
    ```

### Client

The client runs in the customer's environment and establishes a reverse tunnel to the server.

**To run the client:**

Run the client executable, providing the authentication string, server address, and the reverse tunnel remote definition.

The remote `R:9000:db1.customer.internal:5432` tells the server to listen on its port `9000` and forward all traffic to the client, which will then connect to `db1.customer.internal:5432`.

```bash
./sedai-agent-v4 client \
  --auth "customer01:super-secret-token-for-customer01" \
  your-tenant.saas.com:443 \
  R:9000:db1.customer.internal:5432
```
