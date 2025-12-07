# Security Considerations

This document outlines security best practices and considerations for the Cortex Wiki.

## Current Security Measures

### 1. Content Control

- **Publish Flag**: Only content marked with `publish: "true"` in frontmatter is published
- **Source Validation**: Content is synced from a trusted Obsidian vault
- **Safe File Processing**: Python sync script includes proper error handling and validation

### 2. Dependency Management

- Regular dependency updates using `npm update` and `npx @astrojs/upgrade`
- No known vulnerabilities in current dependency tree
- TypeScript for type safety

### 3. Code Quality

- ESLint for static code analysis
- Prettier for consistent code formatting
- TypeScript strict mode enabled

## Recommended Security Enhancements

### Subresource Integrity (SRI)

The site currently loads CSS from an external CDN without SRI hashes:

```html
<link rel="stylesheet" href="https://latex.now.sh/style.css" />
```

**Recommendation**: Add SRI hashes to ensure CDN resources haven't been tampered with.

#### How to Add SRI Hashes

1. **Generate the hash**:

   ```bash
   curl -s https://latex.now.sh/style.css | \
     openssl dgst -sha384 -binary | \
     openssl base64 -A
   ```

2. **Update the link tag** in `src/layouts/BaseLayout.astro`:

   ```html
   <link
     rel="stylesheet"
     href="https://latex.now.sh/style.css"
     integrity="sha384-HASH_HERE"
     crossorigin="anonymous"
   />
   ```

**Note**: SRI hashes need to be updated whenever the external resource changes.

### Content Security Policy (CSP)

Consider adding a Content Security Policy header to prevent XSS attacks.

**Add to `public/_headers`** (for static hosting):

```
/*
  Content-Security-Policy: default-src 'self'; style-src 'self' https://latex.now.sh; script-src 'self'; img-src 'self' data:; font-src 'self'
```

Or configure in your hosting provider's settings.

## Development Security

### Environment Variables

- Never commit `.env` files
- Use environment variables for sensitive configuration
- The `.gitignore` already excludes `.env` files

### Python Script Security

- `obsidian_sync.py` includes:
  - Path validation
  - Error handling
  - Dry-run mode for testing
  - Configurable paths (not hardcoded)

### Dependencies

- Run `npm audit` regularly to check for vulnerabilities
- Use `npm audit fix` to automatically fix issues
- Keep dependencies updated with `make upgrade`

## Deployment Security

### GitHub Actions

- Uses pinned action versions (v4, v2)
- Minimal permissions (read for contents, write for pages)
- No secrets exposed in workflows

### GitHub Pages

- HTTPS enforced via GitHub Pages settings
- Custom domain with proper DNS configuration

## Monitoring

### Regular Checks

- [ ] Run `npm audit` monthly
- [ ] Update dependencies quarterly
- [ ] Review GitHub security alerts
- [ ] Check external CDN resources are still available

### Automated Security

- Dependabot can be enabled for automatic dependency updates
- GitHub Advanced Security can scan for vulnerabilities

## Reporting Security Issues

If you discover a security vulnerability:

1. **Do not** open a public issue
2. Email the repository owner at the address in their GitHub profile
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [MDN Web Security](https://developer.mozilla.org/en-US/docs/Web/Security)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [Astro Security](https://docs.astro.build/en/guides/security/)
