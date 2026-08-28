# AppConfig API Documentation

The AppConfig module serves global application initialization configuration, feature flags, versioning rules, split types, categories, currencies, system limits, payment integrations, legal policy links, and user context with HTTP ETag / `If-None-Match` conditional caching support (`Cache-Control: public, max-age=300`).

---

## Data Models

### AppConfigResponse
```json
{
  "data": {
    "system": {
      "appVersion": {
        "minSupportedVersion": "1.0.0",
        "latestVersion": "1.2.0",
        "forceUpdate": false,
        "updateUrl": {
          "android": "https://play.google.com/store/apps/details?id=com.splittr.app",
          "ios": "https://apps.apple.com/app/splittr/id123456789"
        },
        "updateMessage": "A new version of Splittr is available."
      },
      "maintenance": {
        "inMaintenance": false,
        "readOnlyMode": false,
        "message": "",
        "estimatedEndTime": null
      }
    },
    "domain": {
      "categories": [
        {
          "id": "cat-food",
          "name": "Food & Drink",
          "iconUrl": "https://cdn.splittr.app/icons/food.svg"
        },
        {
          "id": "cat-shopping",
          "name": "Shopping",
          "iconUrl": "https://cdn.splittr.app/icons/shopping.svg"
        }
      ],
      "currencies": [
        {
          "code": "USD",
          "symbol": "$",
          "name": "US Dollar",
          "decimalPlaces": 2,
          "isDefault": true
        },
        {
          "code": "EUR",
          "symbol": "€",
          "name": "Euro",
          "decimalPlaces": 2,
          "isDefault": false
        },
        {
          "code": "INR",
          "symbol": "₹",
          "name": "Indian Rupee",
          "decimalPlaces": 2,
          "isDefault": false
        }
      ],
      "splitTypes": [
        {
          "code": "EQUAL",
          "label": "Split Equally",
          "description": "Split the cost equally among all participants"
        },
        {
          "code": "EXACT",
          "label": "Exact Amounts",
          "description": "Specify the exact amount each participant owes"
        },
        {
          "code": "PERCENTAGE",
          "label": "By Percentage",
          "description": "Specify the percentage share for each participant"
        }
      ],
      "limits": {
        "maxExpenseAmount": 100000.00,
        "maxGroupMembers": 50,
        "maxSplitParticipants": 50,
        "maxReceiptSizeMb": 5,
        "allowedReceiptMimeTypes": [
          "image/jpeg",
          "image/png",
          "image/webp",
          "application/pdf"
        ]
      },
      "paymentIntegrations": [
        {
          "id": "upi",
          "name": "UPI",
          "enabled": true,
          "deepLinkScheme": "upi://pay"
        }
      ]
    },
    "featureFlags": {
      "enableReceiptScanning": true,
      "enableMultiCurrency": true,
      "enableDebtSimplification": true,
      "enableOfflineSync": true
    },
    "legal": {
      "termsOfServiceUrl": "https://splittr.app/terms",
      "privacyPolicyUrl": "https://splittr.app/privacy",
      "faqUrl": "https://splittr.app/faq",
      "supportEmail": "support@splittr.app"
    },
    "userContext": {
      "isAuthenticated": true,
      "userPreferredCurrency": "USD",
      "userFeatureFlags": {}
    }
  },
  "meta": {
    "configVersion": "v1.2.0-hash123",
    "serverTime": "2026-08-28T18:00:00Z"
  }
}
```

---

## Endpoints

### 1. Fetch Application Startup Configuration
- **GET** `/app-config`
- **Authentication**: Optional (`OptionalAuthenticate`) — If a valid `Bearer <TOKEN>` is provided, `data.userContext` will be populated with `isAuthenticated: true` and the user's preferred settings.
- **Request Headers**:
  - `If-None-Match` (string, optional): ETag hash from previous response `meta.configVersion`.
- **Response Headers**:
  - `ETag`: Current configuration version hash.
  - `Cache-Control`: `public, max-age=300`
- **Response Status Codes**:
  - `200 OK`: `AppConfigResponse` (fresh configuration returned).
  - `304 Not Modified`: Config has not changed since the provided `If-None-Match` ETag.
  - `500 Internal Server Error`: `ErrorResponse`.
