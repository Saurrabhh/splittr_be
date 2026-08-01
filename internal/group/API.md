# Group API Documentation

The Group module manages bill-splitting groups, member rosters, invite codes, role permissions (admin vs member), join request approval workflows, and group activity feeds.

## Endpoints

### 1. Create Group
- **POST** `/groups`
- **Description**: Create a new bill-splitting group with optional admin approval for join requests.
- **Request Body**:
  ```json
  {
    "name": "Trip to Japan",
    "description": "Tokyo & Kyoto 2026",
    "requireAdminApproval": true
  }
  ```
- **Response** (`201 Created`): `Group` object.

### 2. List User Groups
- **GET** `/groups?limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of active groups the requesting user belongs to.
- **Response** (`200 OK`): `ListGroupsResponse`.

### 3. Get Group Details
- **GET** `/groups/{id}`
- **Description**: Get metadata and member roster for a specific group.
- **Response** (`200 OK`): `DetailsResponse`.

### 4. Join Group
- **POST** `/groups/join`
- **Description**: Join a group using an active invite code.
- **Request Body**:
  ```json
  {
    "inviteCode": "invite-a1b2c3d4"
  }
  ```
- **Response** (`200 OK`): `JoinResponse` (Status: `ACTIVE` or `PENDING`).

### 5. Get Group Preview
- **GET** `/groups/preview?inviteCode={string}`
- **Description**: Inspect basic group details before joining.
- **Response** (`200 OK`): `Preview`.

### 6. List Group Members
- **GET** `/groups/{id}/members?status={ACTIVE|PENDING|REJECTED|ALL}`
- **Description**: List members of a group. Non-active queries require admin role.
- **Response** (`200 OK`): Array of `Member` objects.

### 7. Add Group Member
- **POST** `/groups/{id}/members`
- **Description**: Add a user directly to the group by User ID (Admin only).
- **Response** (`200 OK`): `MessageResponse`.

### 8. Remove Member / Leave Group
- **DELETE** `/groups/{id}/members/{userId}`
- **Description**: Remove a member from the group or leave the group.
- **Response** (`204 No Content`).

### 9. Update Member Role
- **PUT** `/groups/{id}/members/{userId}/role`
- **Description**: Change a member's role (`admin` or `member`). Admin only.
- **Response** (`200 OK`): `MessageResponse`.

### 10. Decide Join Request
- **POST** `/groups/{id}/members/{userId}/decision`
- **Description**: Approve or reject a pending join request. Admin only.
- **Response** (`200 OK`): `MessageResponse`.

### 11. Reset Invite Code
- **POST** `/groups/{id}/invite-code/reset`
- **Description**: Generate a new invite code with a 7-day expiration. Admin only.
- **Response** (`200 OK`): Updated `Group` object.

### 12. Archive Group
- **DELETE** `/groups/{id}`
- **Description**: Soft-delete a group. Admin only.
- **Response** (`204 No Content`).

### 13. Get Group Activity Feed
- **GET** `/groups/{id}/feed?limit={int}&cursor={string}`
- **Description**: Get cursor-paginated timeline of events inside the group.
- **Response** (`200 OK`): `activity.FeedResponse`.
