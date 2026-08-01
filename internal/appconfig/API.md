# AppConfig API Documentation

The AppConfig module serves global application initialization configuration, feature flags, versioning rules, categories, currencies, limits, and legal policy links with ETag caching support.

## Endpoints

### 1. Fetch Application Startup Configuration
- **GET** `/app-config`
- **Headers**: `If-None-Match: <etag_string>` (optional)
- **Description**: Returns version requirements, feature flags, system limits, active currencies, categories, and maintenance state.
- **Response** (`200 OK` or `304 Not Modified`): `AppConfigResponse`.
