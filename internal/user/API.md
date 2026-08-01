# User API Documentation

The User module manages authentication identity mapping (Firebase UID to local user ID), user profiles, default currencies, and direct peer-to-peer friendship relations.

## Endpoints

### 1. Register User
- **POST** `/users`
- **Description**: Create a local user profile linked to a Firebase authenticated identity.
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

### 4. Add Friend
- **POST** `/friends`
- **Description**: Link another user as a friend using their email or phone number.
- **Request Body**:
  ```json
  {
    "friendEmail": "bob@example.com",
    "friendPhone": ""
  }
  ```
- **Response** (`200 OK`): Friend `User` object.

### 5. List Friends
- **GET** `/friends?limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of the current user's friends.
- **Response** (`200 OK`): `ListFriendsResponse`.

### 6. Remove Friend
- **DELETE** `/friends/{friendId}`
- **Description**: Remove a friendship link.
- **Response** (`204 No Content`).
