# Notification API Documentation

The Notification module manages in-app alerts and notifications generated for users by system actions, group events, or expense settlements.

---

## Data Models

### Notification
```json
{
  "id": "notif-uuid-1",
  "userId": "usr-uuid-1",
  "actorId": "usr-uuid-2",
  "actorName": "Bob Smith",
  "activityId": "act-uuid-1",
  "type": "EXPENSE_ADDED",
  "title": "New Expense Added",
  "content": "Bob Smith added 'Dinner' for $120.00 in Japan Trip 2026",
  "isRead": false,
  "createdAt": "2026-08-02T18:30:00Z"
}
```

### AlertType Enum
- `"EXPENSE_ADDED"`: Notification sent when a user is added to an expense.
- `"PAYMENT_RECEIVED"`: Notification sent to recipient when a settlement payment is recorded.
- `"JOIN_REQUEST_PENDING"`: Notification sent to group admin when a member requests to join.
- `"JOIN_REQUEST_APPROVED"`: Notification sent to member when their join request is approved.
- `"JOIN_REQUEST_REJECTED"`: Notification sent to member when their join request is rejected.

---

## Endpoints

### 1. List Notifications
- **GET** `/notifications?limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve a cursor-paginated list of notifications for the currently logged-in user.
- **Query Parameters**:
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Base64 cursor token from previous response.
- **Response** (`200 OK`): Paginated envelope `pagination.Response[Notification]`.
  ```json
  {
    "data": [
      {
        "id": "notif-uuid-1",
        "userId": "usr-uuid-1",
        "actorId": "usr-uuid-2",
        "actorName": "Bob Smith",
        "activityId": "act-uuid-1",
        "type": "EXPENSE_ADDED",
        "title": "New Expense Added",
        "content": "Bob Smith added 'Dinner' for $120.00 in Japan Trip 2026",
        "isRead": false,
        "createdAt": "2026-08-02T18:30:00Z"
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

### 2. Mark Notification as Read
- **PATCH** `/notifications/{id}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Mark a specific notification as read by its UUID.
- **Path Parameters**:
  - `id` (string, required): Notification UUID.
- **Request Body**:
  ```json
  {
    "isRead": true
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "message": "notification marked as read"
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `404 Not Found` (if not found or belongs to another user), `500 Internal Server Error`.

---

### 3. Mark All Notifications as Read
- **PATCH** `/notifications`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Mark all unread notifications as read for the current user.
- **Request Body**:
  ```json
  {
    "isRead": true
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "message": "all notifications marked as read"
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.
