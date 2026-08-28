# Activity Domain API Specification

The Activity API provides feed and audit log streams for events across Splittr. It features cursor-based pagination and type-safe discriminated JSON payloads for client apps and mobile offline sync integration.

---

## Enums & Data Models

### ActionType & EntityType
- **`entityType`**: `"EXPENSE"` | `"SETTLEMENT"` | `"MEMBER"` | `"GROUP"`
- **`actionType`**:
  - Expense actions: `"EXPENSE_CREATED"`
  - Settlement actions: `"SETTLEMENT"`
  - Member actions: `"MEMBER_ADDED"`, `"MEMBER_LEFT"`, `"MEMBER_KICKED"`, `"MEMBER_ROLE_UPDATED"`, `"MEMBER_JOINED"`
  - Group actions: `"GROUP_CREATED"`, `"GROUP_UPDATED"`, `"GROUP_ARCHIVED"`

### Activity
```json
{
  "id": "c1f7b84a-9b16-4c92-95f2-491d904b77f1",
  "groupId": "3d5f94b1-8b22-4821-bcf3-54bf93f18e10",
  "actor": {
    "id": "usr-uuid-1",
    "name": "Jane Doe"
  },
  "actionType": "EXPENSE_CREATED",
  "entityType": "EXPENSE",
  "entityId": "exp-uuid-1",
  "description": "added expense 'Dinner' of 45.00 USD",
  "payload": {
    "type": "EXPENSE",
    "expense": {
      "id": "exp-uuid-1",
      "description": "Dinner",
      "amount": 45.00,
      "currency": "USD"
    },
    "splits": [ ... ]
  },
  "createdAt": "2026-08-01T20:00:00Z"
}
```

---

## Polymorphic Payload Types (`payload.type`)

The `payload` field is a polymorphic object discriminated by `entityType` and `payload.type`:

### A. `EXPENSE` Payload (`payload.type = "EXPENSE"`)
```json
{
  "type": "EXPENSE",
  "expense": {
    "id": "exp-uuid-1",
    "description": "Dinner at Mario's",
    "amount": 120.00,
    "currency": "USD"
  },
  "splits": [
    {
      "userId": "usr-uuid-1",
      "amount": 60.00
    },
    {
      "userId": "usr-uuid-2",
      "amount": 60.00
    }
  ]
}
```

### B. `SETTLEMENT` Payload (`payload.type = "SETTLEMENT"`)
```json
{
  "type": "SETTLEMENT",
  "expense": {
    "id": "exp-uuid-settle-1",
    "amount": 50.00,
    "currency": "USD"
  },
  "split": {
    "userId": "usr-uuid-1",
    "amount": 50.00
  }
}
```

### C. `MEMBER` Payload (`payload.type = "MEMBER"`)
```json
{
  "type": "MEMBER",
  "member": {
    "userId": "usr-uuid-2",
    "role": "ADMIN",
    "status": "ACTIVE"
  }
}
```

### D. `GROUP` Payload (`payload.type = "GROUP"`)
```json
{
  "type": "GROUP",
  "group": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Japan Trip 2026"
  },
  "members": [ ... ]
}
```

---

## Endpoints

### 1. User Global Activity Feed
- **GET** `/activities?limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Returns cursor-paginated activities visible to the current user (group activities for groups user belongs to + direct non-group user activity events).
- **Query Parameters**:
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Base64 cursor token from previous response.
- **Response** (`200 OK`): Paginated envelope `pagination.Response[Activity]`.
  ```json
  {
    "data": [
      {
        "id": "c1f7b84a-9b16-4c92-95f2-491d904b77f1",
        "groupId": "3d5f94b1-8b22-4821-bcf3-54bf93f18e10",
        "actor": {
          "id": "usr-uuid-1",
          "name": "Jane Doe"
        },
        "actionType": "EXPENSE_CREATED",
        "entityType": "EXPENSE",
        "entityId": "exp-uuid-1",
        "description": "added expense 'Dinner' of 45.00 USD",
        "payload": { ... },
        "createdAt": "2026-08-01T20:00:00Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "hasMore": false,
      "nextCursor": null
    }
  }
  ```
- **Errors**: `401 Unauthorized`, `500 Internal Server Error`.

---

### 2. Group Activity Feed
- **GET** `/groups/{id}/feed?limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Returns cursor-paginated activity log for a specific group. User must be an active group member.
- **Path Parameters**:
  - `id` (string, required): Group UUID.
- **Query Parameters**:
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Base64 cursor token.
- **Response** (`200 OK`): Paginated envelope `pagination.Response[Activity]`.
- **Errors**: `401 Unauthorized`, `403 Forbidden` (not a group member), `404 Not Found`, `500 Internal Server Error`.
