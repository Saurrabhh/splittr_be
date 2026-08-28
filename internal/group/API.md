# Group API Documentation

The Group module manages bill-splitting groups, member rosters, invite codes, role permissions (admin vs member), join request approval workflows, and group activity feeds.

---

## Data Models

### Group
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Japan Trip 2026",
  "description": "Tokyo & Kyoto group expenses",
  "inviteCode": "INV-A1B2C3D4",
  "inviteCodeExpiresAt": "2026-08-09T12:00:00Z",
  "requireAdminApproval": true,
  "createdBy": "usr-creator-uuid",
  "iconUrl": "https://res.cloudinary.com/.../icon.jpg",
  "createdAt": "2026-08-02T12:00:00Z",
  "updatedAt": "2026-08-02T12:00:00Z",
  "archivedAt": null
}
```

### Member
```json
{
  "groupId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "usr-member-uuid",
  "role": "ADMIN",
  "status": "ACTIVE",
  "joinedAt": "2026-08-02T12:00:00Z",
  "name": "Saurabh Yadav",
  "email": "saurabh@example.com",
  "phone": "+1234567890"
}
```
- `role`: `"ADMIN"` | `"MEMBER"`
- `status`: `"ACTIVE"` | `"PENDING"` | `"REJECTED"`

### DetailsResponse
```json
{
  "group": { ... },
  "members": [ { ... } ]
}
```

### Preview
```json
{
  "name": "Japan Trip 2026",
  "description": "Tokyo & Kyoto group expenses",
  "memberCount": 5,
  "creatorName": "Saurabh Yadav",
  "requireAdminApproval": true
}
```

---

## Endpoints

### 1. Create Group
- **POST** `/groups`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Create a new bill-splitting group. The creator is automatically added as an active `ADMIN`.
- **Request Body**:
  ```json
  {
    "name": "Trip to Japan",
    "description": "Tokyo & Kyoto 2026",
    "requireAdminApproval": true
  }
  ```
  - `name` (string, required): Group name.
  - `description` (string, optional): Group description.
  - `requireAdminApproval` (boolean, optional): If true, new members joining via invite code require admin approval (`PENDING` state).
- **Response** (`201 Created`): `Group` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 2. List User Groups
- **GET** `/groups?limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve a cursor-paginated list of active groups the requesting user belongs to.
- **Query Parameters**:
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Opaque URL-safe Base64 cursor token from previous response.
- **Response** (`200 OK`):
  ```json
  {
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Japan Trip 2026",
        "description": "Tokyo & Kyoto group expenses",
        "requireAdminApproval": true,
        "createdAt": "2026-08-02T12:00:00Z",
        "updatedAt": "2026-08-02T12:00:00Z",
        "members": [
          {
            "groupId": "550e8400-e29b-41d4-a716-446655440000",
            "userId": "usr-member-uuid",
            "role": "ADMIN",
            "status": "ACTIVE",
            "name": "Saurabh Yadav"
          }
        ]
      }
    ],
    "pagination": {
      "nextCursor": "MjAyNi0wOC0wMlQxMjowMDowMFpfNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAw",
      "hasMore": true
    }
  }
  ```

- **Errors**: `401 Unauthorized`, `500 Internal Server Error`.

---

### 3. Get Group Details
- **GET** `/groups/{id}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Get metadata and member roster for a specific group. Requester must be an `ACTIVE` member.
- **Path Parameters**:
  - `id` (string, required): Group UUID.
- **Response** (`200 OK`): `DetailsResponse` (`{ "group": {...}, "members": [...] }`).
- **Errors**: `401 Unauthorized`, `403 Forbidden` (not a member), `404 Not Found`, `500 Internal Server Error`.

---

### 4. Update Group
- **PATCH** `/groups/{id}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Update group details (name, description, requireAdminApproval). Admin privileges required.
- **Path Parameters**:
  - `id` (string, required): Group UUID.
- **Request Body**:
  ```json
  {
    "name": "Trip to Japan Updated",
    "description": "Tokyo & Kyoto group expenses",
    "requireAdminApproval": true
  }
  ```
  - `name` (string, required): Group name.
  - `description` (string, optional): Group description.
  - `requireAdminApproval` (boolean, optional): If true, new members joining via invite code require admin approval.
- **Response** (`200 OK`): Updated `Group` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden` (non-admin), `404 Not Found`, `500 Internal Server Error`.

---

### 5. Join Group via Invite Code
- **POST** `/groups/join`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Join an existing group using an active invite code. If `requireAdminApproval` is enabled, membership status will be `PENDING` until an admin approves it.
- **Request Body**:
  ```json
  {
    "inviteCode": "INV-A1B2C3D4"
  }
  ```
- **Response** (`200 OK`): `JoinResponse`
  - Active Join:
    ```json
    {
      "status": "ACTIVE",
      "group": { ... }
    }
    ```
  - Pending Join (requires admin approval):
    ```json
    {
      "status": "PENDING",
      "message": "Join request submitted for admin approval."
    }
    ```
- **Errors**: `400 Bad Request` (expired/invalid code), `401 Unauthorized`, `403 Forbidden` (previously rejected), `500 Internal Server Error`.

---

### 6. Get Group Preview
- **GET** `/groups/preview?inviteCode={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Inspect basic group details before joining via invite code.
- **Query Parameters**:
  - `inviteCode` (string, required): Active group invite code.
- **Response** (`200 OK`): `Preview` object.
- **Errors**: `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`.

---

### 7. List Group Members
- **GET** `/groups/{id}/members?status={ACTIVE|PENDING|REJECTED|ALL}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: List members of a group. Querying status other than `ACTIVE` (or omitting status) requires `ADMIN` role.
- **Query Parameters**:
  - `status` (string, optional): Filter by member status (`ACTIVE`, `PENDING`, `REJECTED`, `ALL`). Default: `ACTIVE`.
- **Response** (`200 OK`): Array of `Member` objects.
- **Errors**: `401 Unauthorized`, `403 Forbidden` (non-admin querying non-active members), `500 Internal Server Error`.

---

### 8. Add Group Members Directly (Bulk)
- **POST** `/groups/{id}/members`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Add multiple users directly to the group by User IDs as `ACTIVE` `MEMBER`s. Admin privileges required.
- **Request Body**:
  ```json
  {
    "userIds": [
      "usr-target-uuid-1",
      "usr-target-uuid-2"
    ]
  }
  ```
- **Response** (`201 Created`): Array of created `Member` objects.
  ```json
  [
    {
      "groupId": "550e8400-e29b-41d4-a716-446655440000",
      "userId": "usr-target-uuid-1",
      "role": "MEMBER",
      "status": "ACTIVE",
      "joinedAt": "2026-08-02T14:30:00Z",
      "name": "John Doe",
      "email": "john@example.com"
    }
  ]
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden` (non-admin), `404 Not Found`, `500 Internal Server Error`.

---

### 9. Remove Member / Leave Group
- **DELETE** `/groups/{id}/members/{userId}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Remove a member from the group (Admin only), or leave the group (if `userId` equals requesting user ID). Sole admin cannot leave without assigning another admin.
- **Path Parameters**:
  - `id` (string, required): Group UUID.
  - `userId` (string, required): Target user UUID to remove or self user ID.
- **Response** (`204 No Content`).
- **Errors**: `400 Bad Request` (sole admin removal error), `401 Unauthorized`, `403 Forbidden` (non-admin removing others), `404 Not Found`, `500 Internal Server Error`.

---

### 10. Update Member Role
- **PATCH** `/groups/{id}/members/{userId}/role`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Update a member's role (`ADMIN` or `MEMBER`). Admin privileges required.
- **Request Body**:
  ```json
  {
    "role": "ADMIN"
  }
  ```
- **Response** (`200 OK`): Updated `Member` object.
  ```json
  {
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "usr-target-uuid",
    "role": "ADMIN",
    "status": "ACTIVE",
    "joinedAt": "2026-08-02T14:30:00Z",
    "name": "John Doe"
  }
  ```
- **Errors**: `400 Bad Request` (invalid role), `401 Unauthorized`, `403 Forbidden` (non-admin), `404 Not Found`, `500 Internal Server Error`.

---

### 11. Decide Join Request
- **POST** `/groups/{id}/members/{userId}/decision`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Approve (`APPROVE`) or reject (`REJECT`) a pending member join request. Admin privileges required.
- **Request Body**:
  ```json
  {
    "action": "APPROVE"
  }
  ```
  - `action` (string, required): `"APPROVE"` or `"REJECT"`.
- **Response** (`200 OK`): Updated `Member` object with updated status (`ACTIVE` or `REJECTED`).
  ```json
  {
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "usr-target-uuid",
    "role": "MEMBER",
    "status": "ACTIVE",
    "joinedAt": "2026-08-02T14:30:00Z",
    "name": "Jane Smith"
  }
  ```
- **Errors**: `400 Bad Request` (invalid action), `401 Unauthorized`, `403 Forbidden` (non-admin), `404 Not Found`, `500 Internal Server Error`.

---

### 12. Reset Invite Code
- **POST** `/groups/{id}/invite-code/reset`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Generate a fresh 8-character invite code (`INV-XXXXXXXX`) with a 7-day expiration timestamp. Admin privileges required.
- **Response** (`200 OK`): Updated `Group` object.
- **Errors**: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `500 Internal Server Error`.

---

### 13. Archive Group
- **DELETE** `/groups/{id}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Soft-delete a group (`archivedAt` timestamp set). Admin privileges required.
- **Response** (`204 No Content`).
- **Errors**: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `500 Internal Server Error`.

---

### 14. Get Group Activity Feed
- **GET** `/groups/{id}/feed?limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve a cursor-paginated timeline of events inside the group (expense created, settlements, member joined, etc.). Requester must be an `ACTIVE` member.
- **Query Parameters**:
  - `limit` (int, optional): Items per page (default 20, max 100).
  - `cursor` (string, optional): Cursor token from previous response.
- **Response** (`200 OK`): Paginated `Activity` response (`{ "data": [...], "pagination": {...} }`).
- **Errors**: `401 Unauthorized`, `403 Forbidden` (not a group member), `500 Internal Server Error`.

---

### 15. Upload Group Icon
- **POST** `/groups/{id}/icon`
- **Authentication**: Required (`BearerAuth`)
- **Content-Type**: `multipart/form-data`
- **Description**: Upload an image file (JPEG, PNG, WEBP, max 2MB) as group icon. Automatically resizes/stores in Cloudinary and updates `groups.icon_url`. Requester must be an active group member.
- **Path Parameters**:
  - `id` (string, required): Group UUID.
- **Form Data**:
  - `file` (file, required): Binary image file.
- **Response** (`200 OK`): Updated `Group` object.
- **Errors**: `400 Bad Request` (invalid format/missing file), `401 Unauthorized`, `403 Forbidden` (not an active group member), `500 Internal Server Error`.



