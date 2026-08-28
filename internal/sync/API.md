# Unified Sync API Documentation

The Sync module provides a single, high-performance batch synchronization endpoint (`GET /v1/sync`) for offline-first mobile applications (e.g. Flutter with Drift/Isar). It aggregates delta updates and entity tombstones across Friends, Groups, and Expenses in a single atomic round-trip.

---

## Data Models

### SyncResponse
```json
{
  "friends": {
    "newVersion": 150,
    "updated": [
      {
        "userId": "usr-1",
        "friendId": "usr-2",
        "status": "ACCEPTED",
        "actionUserId": "usr-1",
        "createdAt": "2026-08-09T20:00:00Z",
        "syncVersion": 150
      }
    ],
    "deletedIds": [ "usr-3" ]
  },
  "groups": {
    "newVersion": 150,
    "updated": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Japan Trip 2026",
        "description": "Tokyo & Kyoto group expenses",
        "requireAdminApproval": true,
        "createdAt": "2026-08-02T12:00:00Z",
        "updatedAt": "2026-08-02T12:00:00Z",
        "syncVersion": 150,
        "members": [ ... ]
      }
    ],
    "deletedIds": [ "grp-old-uuid" ]
  },
  "expenses": {
    "newVersion": 150,
    "updated": [
      {
        "id": "exp-1",
        "description": "Dinner",
        "amount": 120.00,
        "currency": "USD",
        "category": "Food",
        "groupId": "550e8400-e29b-41d4-a716-446655440000",
        "paidBy": "usr-1",
        "splitType": "EQUAL",
        "splits": [ ... ],
        "syncVersion": 150,
        "createdAt": "2026-08-02T12:00:00Z",
        "updatedAt": "2026-08-02T12:00:00Z"
      }
    ],
    "deletedIds": [ "exp-deleted-uuid" ]
  }
}
```

---

## Endpoints

### 1. Unified Batch Sync
- **GET** `/sync?friendsVersion={int64}&groupsVersion={int64}&expensesVersion={int64}&limit={int}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve delta changes and deletions for friends, groups, and expenses modified after the given sequence versions.
- **Query Parameters**:
  - `friendsVersion` (int64, optional): Last stored sequence version for friends. Default: 0.
  - `groupsVersion` (int64, optional): Last stored sequence version for groups. Default: 0.
  - `expensesVersion` (int64, optional): Last stored sequence version for expenses. Default: 0.
  - `limit` (int, optional): Max items returned per category. Default: 100.
- **Response** (`200 OK`): `SyncResponse` object.
- **Errors**: `401 Unauthorized`, `500 Internal Server Error`.
