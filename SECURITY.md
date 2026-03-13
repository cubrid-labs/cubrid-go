# Security Policy

## Supported Versions

| Version | Status |
|---------|--------|
| 0.1.x   | ✅ Supported |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue in cubrid-go, please report it responsibly by emailing:

**Email:** paikend@gmail.com

**Do not** open a public GitHub issue for security vulnerabilities. Responsible disclosure allows us to address the issue before public disclosure.

### Response Timeline

- **48 hours:** Initial acknowledgment of your report
- **7 days:** Security assessment and initial response with remediation plan
- **Ongoing:** Regular updates on progress until resolution

## What Qualifies as a Security Issue

A security issue is any vulnerability that could:

- Allow unauthorized access to data
- Permit SQL injection or other code execution attacks
- Expose sensitive information (credentials, tokens, private data)
- Compromise confidentiality, integrity, or availability of the system

Examples include:
- SQL injection vulnerabilities in query construction
- Authentication/authorization flaws
- Insecure credential handling
- Buffer overflow or memory safety issues in protocol handling

## Security Best Practices for Users

When using cubrid-go, follow these security best practices:

- Always use parameterized queries to prevent SQL injection
- Keep cubrid-go updated to the latest version
- Use secure connection parameters when connecting to CUBRID databases
- Follow the principle of least privilege for database credentials
- Never hardcode credentials in your application code
- Use environment variables or secure credential management systems

## Disclosure Policy

Once a security vulnerability is fixed:

1. A security patch will be released
2. The vulnerability will be disclosed in release notes
3. Credit will be given to the reporter (if requested)

We appreciate your responsible disclosure and help in keeping cubrid-go secure.
