# User API Documentation

The User module manages authentication identity mapping (Firebase UID to local user ID), user profiles, avatars, default currencies, user preferences (`user_settings`), and stateful peer-to-peer friendship relations (`PENDING`, `ACCEPTED`, `DECLINED`, `BLOCKED`).

---

## Data Models

### User
```json
{
  "id": "usr-uuid-1",
  "name": "Alice Smith",
  "email": "alice@example.com",
  "phone": "+1234567890",
  "defaultCurrency": "USD",
  "avatarUrl": "https://res.cloudinary.com/.../avatar.jpg",
  "createdAt": "2026-08-01T12:00:00Z",
  "updatedAt": "2026-08-01T12:00:00Z"
}
```

### UserSettingsResponse
```json
{
  "autoAcceptFriendRequests": false
}
```

### FriendWithStatus
```json
{
  "id": "usr-uuid-2",
  "name": "Bob",
  "email": "bob@example.com",
  "phone": "+1234567890",
  "defaultCurrency": "INR",
  "avatarUrl": "https://res.cloudinary.com/.../bob.jpg",
  "status": "PENDING",
  "actionUserId": "usr-uuid-1",
  "createdAt": "2026-08-14T15:45:00Z",
  "updatedAt": "2026-08-14T15:45:00Z"
}
```
- `status`: `"PENDING"` | `"ACCEPTED"` | `"DECLINED"` | `"BLOCKED"`
- `actionUserId`: UUID of the user who performed the last state transition (e.g. initiator of the friend request or user who accepted/blocked).

---

## Endpoints

### 1. Register User
- **POST** `/users`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Create a local user profile linked to a Firebase authenticated identity and initialize default user settings (`autoAcceptFriendRequests: false`).
- **Request Body**:
  ```json
  {
    "name": "Alice Smith"
  }
  ```
- **Response** (`201 Created`): `User` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 2. Get Current User Profile
- **GET** `/users/me`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve the profile details of the currently logged-in user.
- **Response** (`200 OK`): `User` object.
- **Errors**: `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`.

---

### 3. Update User Profile
- **PATCH** `/users/me`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Update the name or default currency of the current user.
- **Request Body**:
  ```json
  {
    "name": "Alice Vance",
    "defaultCurrency": "EUR"
  }
  ```
- **Response** (`200 OK`): Updated `User` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 4. Upload Profile Avatar
- **POST** `/users/me/avatar`
- **Authentication**: Required (`BearerAuth`)
- **Content-Type**: `multipart/form-data`
- **Description**: Upload an image file (JPEG, PNG, WEBP, max 2MB) as user avatar. Automatically resizes/stores in Cloudinary and updates `avatarUrl`.
- **Form Data**:
  - `file` (file, required): Binary image file.
- **Response** (`200 OK`): Updated `User` object with new `avatarUrl`.
- **Errors**: `400 Bad Request` (unsupported type/missing file), `401 Unauthorized`, `500 Internal Server Error`.

---

### 5. Get User Settings
- **GET** `/users/me/settings`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve privacy and application preferences for the current user.
- **Response** (`200 OK`): `UserSettingsResponse` (`{ "autoAcceptFriendRequests": false }`).
- **Errors**: `401 Unauthorized`, `500 Internal Server Error`.

---

### 6. Update User Settings
- **PATCH** `/users/me/settings`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Update user preferences (e.g. toggle auto-accept for incoming friend requests).
- **Request Body**:
  ```json
  {
    "autoAcceptFriendRequests": true
  }
  ```
- **Response** (`200 OK`): Updated `UserSettingsResponse`.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 7. Add Friend / Send Friend Request
- **POST** `/friends`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Create a friendship link with another user by email or phone. If target user has `autoAcceptFriendRequests: true`, status is set to `ACCEPTED` immediately; otherwise `PENDING`.
- **Request Body**:
  ```json
  {
    "friendEmail": "bob@example.com",
    "friendPhone": ""
  }
  ```
- **Response** (`200 OK`): `FriendWithStatus` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `409 Conflict`, `500 Internal Server Error`.

---

### 8. List Friends
- **GET** `/friends?limit={int}&cursor={string}&status={ACCEPTED|PENDING|BLOCKED}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve friends list. Defaults to cursor-paginated list of accepted friends using URL-safe Base64 cursor tokens. If `status` query parameter is supplied, returns matching friends with status metadata (`actionUserId`, `status`).
- **Query Parameters**:
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Base64 cursor token.
  - `status` (string, optional): Filter by friendship status (`ACCEPTED`, `PENDING`, `BLOCKED`).
- **Response** (`200 OK`): Array/page of `FriendWithStatus` objects.
- **Errors**: `401 Unauthorized`, `500 Internal Server Error`.

---

### 9. Update Friendship Status (Accept / Decline / Block)
- **PATCH** `/friends/{friendId}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Update status of a friendship or friend request (`ACCEPTED`, `DECLINED`, `BLOCKED`).
- **Path Parameters**:
  - `friendId` (string, required): Target user UUID.
- **Request Body**:
  ```json
  {
    "status": "ACCEPTED"
  }
  ```
- **Response** (`200 OK`): Updated `FriendWithStatus` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`.

---

### 10. Remove Friend
- **DELETE** `/friends/{friendId}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Remove a friendship link. Generates tombstones to propagate deletion to offline clients via delta sync.
- **Path Parameters**:
  - `friendId` (string, required): Friend user UUID.
- **Response** (`204 No Content`).
- **Errors**: `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`.
