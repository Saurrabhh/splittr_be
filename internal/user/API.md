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
- **PUT** `/users/me`
- **Description**: Update the name or default currency of the current user.
- **Request Body**:
  ```json
  {
    "name": "Alice Vance",
    "defaultCurrency": "EUR"
  }
  ```
- **Response** (`200 OK`): Updated `User` object.

### 4. Get User Settings
- **GET** `/users/me/settings`
- **Description**: Retrieve privacy and application preferences for the current user.
- **Response** (`200 OK`):
  ```json
  {
    "autoAcceptFriendRequests": false
  }
  ```

### 5. Update User Settings
- **PUT** `/users/me/settings`
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
- **Response** (`200 OK`):
  ```json
  {
    "friend": {
      "id": "usr-uuid-2",
      "name": "Bob",
      "email": "bob@example.com",
      "defaultCurrency": "INR"
    },
    "status": "PENDING"
  }
  ```

### 7. List Friends
- **GET** `/friends?limit={int}&cursor={string}&status={ACCEPTED|PENDING|BLOCKED}`
- **Description**: Retrieve friends list. Defaults to cursor-paginated list of accepted friends. If `status` query parameter is supplied, returns matching friends with status metadata (`actionUserId`, `status`).
- **Response** (`200 OK`): Paginated or array list of friends.

### 8. Update Friendship Status (Accept / Decline / Block)
- **PATCH** `/friends/{friendId}`
- **Description**: Update status of a friendship or friend request (`ACCEPTED`, `DECLINED`, `BLOCKED`).
- **Request Body**:
  ```json
  {
    "status": "ACCEPTED"
  }
  ```
- **Response** (`204 No Content`).

### 9. Remove Friend
- **DELETE** `/friends/{friendId}`
- **Description**: Remove a friendship link.
- **Response** (`204 No Content`).

### 10. Sync Friends
- **GET** `/friends/sync?lastVersion={int64}&limit={int}`
- **Description**: Retrieve friendship changes (including status and `actionUserId`) modified after a given sequence version for offline sync.
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
    ]
  }
  ```
