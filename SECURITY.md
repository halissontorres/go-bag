# Security Policy

## Overview

The **go-bag** project takes security seriously. We appreciate the efforts of security researchers and the community in helping us maintain a secure and reliable library.

This document outlines how to report vulnerabilities, supported versions, and our disclosure process.

---

## Supported versions

We follow a rolling release model. Only the most recent version is actively maintained with security updates.

| Version | Supported |
|---------|----------|
| latest  | ✅ Yes   |
| older   | ❌ No    |

**Recommendation:** Always upgrade to the latest version to receive security patches.

---

## Reporting a vulnerability

If you believe you have found a security vulnerability, please report it responsibly.

### Preferred method (GitHub Security Advisory)

1. Go to the repository's **Security** tab
2. Click on **"Report a vulnerability"**
3. Submit a private advisory with details

---

### ⚠️ Important

- **Do NOT** open public issues for security vulnerabilities
- Avoid disclosing the vulnerability publicly until it has been addressed

---

## What to include

Please provide as much detail as possible:

- Type of vulnerability (e.g., DoS, data corruption, race condition)
- Affected components or functions
- Step-by-step reproduction instructions
- Proof of Concept (PoC), if available
- Potential impact and attack scenario
- Suggested mitigation or fix (optional)

---

## Response and disclosure timeline

We aim to follow these guidelines:

| Phase                    | Target Time        |
|--------------------------|-------------------|
| Initial acknowledgment   | ≤ 72 hours        |
| Triage and validation    | ≤ 7 days          |
| Fix development          | Depends on severity |
| Coordinated disclosure   | After fix release |

For critical vulnerabilities, we may accelerate this timeline.

---

## Severity assessment

We generally align severity with industry standards such as:

- CVSS (Common Vulnerability Scoring System)
- OWASP Risk Rating Methodology

Severity levels:

- **Critical** – Remote exploitation, major impact
- **High** – Significant security risk
- **Medium** – Limited impact or complex exploitation
- **Low** – Minor issues or hard-to-exploit cases

---

## Security best practices

While **go-bag** is a data structure and utility library, consumers should consider:

- Validate all untrusted input before usage
- Avoid unsafe concurrent access without proper synchronization
- Monitor memory usage in high-load scenarios
- Keep dependencies up to date
- Use static analysis tools (e.g., `go vet`, `gosec`)

---

## Dependency security

We strive to:

- Keep dependencies minimal and audited
- Regularly review and update third-party libraries
- Use automated tools such as Dependabot

If a vulnerability is found in a dependency, please report it upstream as well.

---

## Disclosure policy

- Vulnerabilities will be disclosed publicly only after a fix is available
- We follow **coordinated disclosure**
- Reporters may request anonymity
- Proper credit will be given when appropriate

---

## Security updates

Security fixes will be:

- Released as soon as possible
- Clearly documented in release notes
- Tagged for visibility (e.g., `security`, `CVE` if applicable)

---

## Out of scope

The following are generally out of scope:

- Issues caused by incorrect or unsafe usage of the library
- Vulnerabilities in third-party dependencies (unless directly exploitable via this project)
- Denial of service scenarios caused by unrealistic workloads

---

## Compliance & standards

This project aligns with general best practices from:

- OWASP Secure Coding Practices
- GitHub Security Best Practices
- Go Security Guidelines

---

## Acknowledgements

We thank all contributors and security researchers who responsibly disclose vulnerabilities and help improve the security of this project.

---

## Contact

For any security-related concerns:

- GitHub Security Advisories (preferred)
- Email: halisson[dot]torres[at]gmail[dot]com

---
