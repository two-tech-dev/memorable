# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability in Memorable, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, email **security@two-tech.dev** with:

- A description of the vulnerability
- Steps to reproduce the issue
- The potential impact
- Any suggested fix (optional)

We will acknowledge receipt within 48 hours and aim to provide a fix within 7 days for critical issues.

## Security Considerations

### API Keys

- Never commit API keys to version control
- Use environment variables or the `${ENV_VAR}` syntax in config files
- The `memorable.yaml` config file is listed in `.gitignore` by default

### Database

- Use strong passwords for PostgreSQL in production
- Enable SSL (`sslmode=require`) for remote database connections
- Restrict network access to the database port

### MCP Transport

- The stdio transport communicates only with the parent process
- No network ports are opened in stdio mode
- Future HTTP transport will require authentication configuration
