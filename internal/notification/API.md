# Notification API Documentation

The Notification module manages in-app alerts and notifications generated for users by system actions, group events, or expense settlements.

## Endpoints

### 1. List Notifications
- **GET** `/notifications?limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of notifications for the current user.
- **Response** (`200 OK`): `ListNotificationsResponse`.

### 2. Mark Notification as Read
- **PATCH** `/notifications/{id}`
- **Description**: Mark a specific notification as read by ID.
- **Request Body**:
  ```json
  {
    "isRead": true
  }
  ```
- **Response** (`200 OK`): `MessageResponse`.
- **Response** (`404 Not Found`): `ErrorResponse` — if the notification does not exist or belongs to another user.

### 3. Mark All Notifications as Read
- **PATCH** `/notifications`
- **Description**: Mark all unread notifications as read for the current user.
- **Request Body**:
  ```json
  {
    "isRead": true
  }
  ```
- **Response** (`200 OK`): `MessageResponse`.
