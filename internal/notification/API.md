# Notification API Documentation

The Notification module manages in-app alerts and notifications generated for users by system actions, group events, or expense settlements.

## Endpoints

### 1. List Notifications
- **GET** `/notifications?limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of notifications for the current user.
- **Response** (`200 OK`): `ListNotificationsResponse`.

### 2. Mark Notification as Read
- **POST** `/notifications/{id}/read`
- **Description**: Mark a specific notification as read by ID.
- **Response** (`200 OK`): `MessageResponse`.

### 3. Mark All Notifications as Read
- **POST** `/notifications/read-all`
- **Description**: Mark all unread notifications as read for the current user.
- **Response** (`200 OK`): `MessageResponse`.
