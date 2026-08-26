# User API Documentation

The User module manages authentication identity mapping (Firebase UID to local user ID), user profiles, default currencies, user preferences (`user_settings`), and stateful peer-to-peer friendship relations (`PENDING`, `ACCEPTED`, `DECLINED`, `BLOCKED`).

## Endpoints

### 1. Register User
- **POST** `/users`
- **Description**: Create a local user profile linked to a Firebase authenticated identity and initialize default user settings (`autoAcceptFriendRequests: false`).
- **Request Body**:
  ```json
  {
    "name": "Alice Smith"
  }
  ```
- **Response** (`201 Created`): `User` object.

### 2. Get Current User Profile
- **GET** `/users/me`
- **Description**: Retrieve the profile details of the currently logged-in user.
- **Response** (`200 OK`): `User` object.

### 3. Update User Profile
- **PATCH** `/users/me`
- **Description**: Update the name or default currency of the current user.
- **Request Body**:
  ```json
  {
    "name": "Alice Vance",
    "defaultCurrency": "EUR"
  }
  ```
- **Response** (`200 OK`): Updated `User` object.

### 4. Upload Profile Avatar
- **POST** `/users/me/avatar`
- **Authentication**: Required (`BearerAuth`)
- **Content-Type**: `multipart/form-data`
- **Description**: Upload an image file (JPEG, PNG, WEBP, max 2MB) as user avatar. Automatically resizes/stores in Cloudinary and updates `avatarUrl`.
- **Form Data**:
  - `file` (file, required): Binary image file.
- **Response** (`200 OK`):
  ```json
  {
    "id": "usr-uuid-1",
    "name": "Alice Vance",
    "email": "alice@example.com",
    "phone": "+1234567890",
    "defaultCurrency": "EUR",
    "avatarUrl": "https://res.cloudinary.com/.../avatar.jpg",
    "createdAt": "2026-08-02T12:00:00Z",
    "updatedAt": "2026-08-14T12:00:00Z"
  }
  ```
- **Errors**: `400 Bad Request` (unsupported type/missing file), `401 Unauthorized`, `500 Internal Server Error`.

### 5. Get User Settings
- **GET** `/users/me/settings`
- **Description**: Retrieve privacy and application preferences for the current user.
- **Response** (`200 OK`):
  ```json
  {
    "autoAcceptFriendRequests": false
  }
  ```

### 5. Update User Settings
- **PATCH** `/users/me/settings`
- **Description**: Update user preferences (e.g. toggle auto-accept for incoming friend requests).
- **Request Body**:
  ```json
  {
    "autoAcceptFriendRequests": true
  }
  ```
- **Response** (`200 OK`): Updated `UserSettingsResponse`.

### 6. Add Friend / Send Friend Request
- **POST** `/friends`
- **Description**: Create a friendship link with another user by email or phone. If the target user has `autoAcceptFriendRequests: true`, status is set to `ACCEPTED` immediately; otherwise `PENDING`.
- **Request Body**:
  ```json
  {
    "friendEmail": "bob@example.com",
    "friendPhone": ""
  }
  ```
- **Response** (`200 OK`): `FriendWithStatus` object.
  ```json
  {
    "id": "usr-uuid-2",
    "name": "Bob",
    "email": "bob@example.com",
    "phone": "+1234567890",
    "defaultCurrency": "INR",
    "createdAt": "2026-08-14T15:45:00Z",
    "updatedAt": "2026-08-14T15:45:00Z",
    "status": "PENDING",
    "actionUserId": "usr-uuid-1"
  }
  ```

### 7. List Friends
- **GET** `/friends?limit={int}&cursor={string}&status={ACCEPTED|PENDING|BLOCKED}`
- **Description**: Retrieve friends list. Defaults to cursor-paginated list of accepted friends using URL-safe Base64 cursor tokens. If `status` query parameter is supplied, returns matching friends with status metadata (`actionUserId`, `status`).
- **Response** (`200 OK`): Paginated or array list of friends.

### 8. Update Friendship Status (Accept / Decline / Block)
- **PATCH** `/friends/{friendId}`
- **Description**: Update status of a friendship or friend request (`ACCEPTED`, `DECLINED`, `BLOCKED`). Returns the updated friend profile and status.
- **Request Body**:
  ```json
  {
    "status": "ACCEPTED"
  }
  ```
- **Response** (`200 OK`): Updated `FriendWithStatus` object.
  ```json
  {
    "id": "usr-uuid-2",
    "name": "Bob",
    "email": "bob@example.com",
    "phone": "+1234567890",
    "defaultCurrency": "INR",
    "createdAt": "2026-08-14T15:45:00Z",
    "updatedAt": "2026-08-14T15:47:00Z",
    "status": "ACCEPTED",
    "actionUserId": "usr-uuid-1"
  }
  ```

### 9. Remove Friend
- **DELETE** `/friends/{friendId}`
- **Description**: Remove a friendship link. Generates tombstones to propagate deletion to offline clients via delta sync.
- **Response** (`204 No Content`).

### 10. Sync Friends
- **GET** `/friends/sync?lastVersion={int64}&limit={int}`
- **Description**: Retrieve friendship changes (including status and `actionUserId`) and removed friend tombstones modified after a given sequence version for offline sync.
- **Response** (`200 OK`): `FriendSyncResponse`
  ```json
  {
    "newVersion": 150,
    "friends": [
      {
        "userId": "usr-1",
        "friendId": "usr-2",
        "status": "ACCEPTED",
        "actionUserId": "usr-1",
        "createdAt": "2026-08-09T20:00:00Z",
        "syncVersion": 150
      }
    ],
    "removedFriendIds": [ "usr-uuid-3" ]
  }
  ```

