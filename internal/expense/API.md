# Expense API Documentation

The Expense module manages shared group and individual expenses, splitting logic (`EQUAL`, `EXACT`, `PERCENTAGE`), balance calculations, debt simplification, and settlement payments.

---

## Data Models

### Expense
```json
{
  "id": "exp-uuid-1",
  "description": "Dinner at Mario's",
  "amount": 120.00,
  "currency": "USD",
  "category": "FOOD_AND_DRINK",
  "groupId": "550e8400-e29b-41d4-a716-446655440000",
  "paidBy": "usr-uuid-1",
  "createdBy": "usr-uuid-1",
  "isPayment": false,
  "spentAt": "2026-08-02T18:00:00Z",
  "createdAt": "2026-08-02T18:00:00Z",
  "updatedAt": "2026-08-02T18:00:00Z",
  "syncVersion": 150
}
```

### Split
```json
{
  "expenseId": "exp-uuid-1",
  "userId": "usr-uuid-2",
  "amount": 60.00,
  "splitType": "EQUAL",
  "splitValue": null,
  "name": "Bob Smith",
  "email": "bob@example.com",
  "phone": "+1234567890"
}
```

### ExpenseWithSplits (Unified Response)
```json
{
  "id": "exp-uuid-1",
  "description": "Dinner at Mario's",
  "amount": 120.00,
  "currency": "USD",
  "category": "FOOD_AND_DRINK",
  "groupId": "550e8400-e29b-41d4-a716-446655440000",
  "paidBy": "usr-uuid-1",
  "createdBy": "usr-uuid-1",
  "isPayment": false,
  "spentAt": "2026-08-02T18:00:00Z",
  "createdAt": "2026-08-02T18:00:00Z",
  "updatedAt": "2026-08-02T18:00:00Z",
  "syncVersion": 150,
  "splits": [
    {
      "expenseId": "exp-uuid-1",
      "userId": "usr-uuid-1",
      "amount": 60.00,
      "splitType": "EQUAL",
      "name": "Alice Vance"
    },
    {
      "expenseId": "exp-uuid-1",
      "userId": "usr-uuid-2",
      "amount": 60.00,
      "splitType": "EQUAL",
      "name": "Bob Smith"
    }
  ]
}
```

### BalanceResponse
```json
{
  "balances": [
    {
      "userId": "usr-uuid-1",
      "userName": "Alice Vance",
      "netBalance": 60.00
    },
    {
      "userId": "usr-uuid-2",
      "userName": "Bob Smith",
      "netBalance": -60.00
    }
  ],
  "settlements": [
    {
      "fromUserId": "usr-uuid-2",
      "fromUserName": "Bob Smith",
      "toUserId": "usr-uuid-1",
      "toUserName": "Alice Vance",
      "amount": 60.00
    }
  ]
}
```

---

## Endpoints

### 1. Create Expense
- **POST** `/expenses`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Log a new expense and calculate split distribution according to the specified `splitType`.
- **Request Body (Equal Split)**:
  ```json
  {
    "description": "Dinner at Mario's",
    "amount": 120.00,
    "currency": "USD",
    "category": "FOOD_AND_DRINK",
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "paidBy": "usr-uuid-1",
    "splitType": "EQUAL",
    "splits": [
      { "userId": "usr-uuid-1" },
      { "userId": "usr-uuid-2" }
    ]
  }
  ```
- **Request Body (Exact Amounts Split)**:
  ```json
  {
    "description": "Grocery run",
    "amount": 100.00,
    "currency": "USD",
    "category": "SHOPPING",
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "paidBy": "usr-uuid-1",
    "splitType": "EXACT",
    "splits": [
      { "userId": "usr-uuid-1", "amount": 70.00 },
      { "userId": "usr-uuid-2", "amount": 30.00 }
    ]
  }
  ```
- **Request Body (Percentage Split)**:
  ```json
  {
    "description": "Hotel Suite",
    "amount": 300.00,
    "currency": "USD",
    "category": "UTILITIES",
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "paidBy": "usr-uuid-1",
    "splitType": "PERCENTAGE",
    "splits": [
      { "userId": "usr-uuid-1", "percentage": 60.0 },
      { "userId": "usr-uuid-2", "percentage": 40.0 }
    ]
  }
  ```
- **Response** (`201 Created`): `ExpenseWithSplits` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 2. Settle Up (Record Debt Payment)
- **POST** `/expenses/settle`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Record a payment transaction between two users to clear or reduce debt.
- **Request Body**:
  ```json
  {
    "amount": 60.00,
    "currency": "USD",
    "groupId": "550e8400-e29b-41d4-a716-446655440000",
    "paidBy": "usr-uuid-2",
    "receivedBy": "usr-uuid-1"
  }
  ```
  - `amount` (float, required): Payment amount.
  - `currency` (string, required): Currency code.
  - `groupId` (string, optional): Group context if settling within a specific group.
  - `paidBy` (string, optional): User ID sending the payment (defaults to authenticated user).
  - `receivedBy` (string, required): User ID receiving the payment.
- **Response** (`201 Created`): `ExpenseWithSplits` object with `isPayment: true`.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

---

### 3. List Expenses
- **GET** `/expenses?groupId={id}&personal={bool}&friendId={id}&limit={int}&cursor={string}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve a cursor-paginated list of expenses with splits. Filter parameters (`groupId`, `personal`, `friendId`) are mutually exclusive — passing more than one returns `400 Bad Request`.
- **Query Parameters**:
  - `groupId` (string, optional): Filter by group UUID.
  - `personal` (boolean, optional): `true` to filter non-group personal expenses.
  - `friendId` (string, optional): Filter direct non-group expenses between current user and a friend.
  - `limit` (int, optional): Page size (default 20, max 100).
  - `cursor` (string, optional): Base64 cursor token from previous response.
- **Response** (`200 OK`): Paginated envelope `pagination.Response[ExpenseWithSplits]`.
  ```json
  {
    "data": [
      {
        "id": "exp-uuid-1",
        "description": "Dinner at Mario's",
        "amount": 120.00,
        "currency": "USD",
        "category": "FOOD_AND_DRINK",
        "groupId": "550e8400-e29b-41d4-a716-446655440000",
        "paidBy": "usr-uuid-1",
        "createdBy": "usr-uuid-1",
        "isPayment": false,
        "spentAt": "2026-08-02T18:00:00Z",
        "createdAt": "2026-08-02T18:00:00Z",
        "updatedAt": "2026-08-02T18:00:00Z",
        "syncVersion": 150,
        "splits": [ ... ]
      }
    ],
    "pagination": {
      "limit": 20,
      "hasMore": false,
      "nextCursor": null
    }
  }
  ```
- **Errors**: `400 Bad Request` (missing or conflicting filters), `401 Unauthorized`, `500 Internal Server Error`.

---

### 4. Get Expense Details
- **GET** `/expenses/{id}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Retrieve a specific expense record and its full split breakdown.
- **Path Parameters**:
  - `id` (string, required): Expense UUID.
- **Response** (`200 OK`): `ExpenseWithSplits` object.
- **Errors**: `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `500 Internal Server Error`.

---

### 5. Update Expense
- **PATCH** `/expenses/{id}`
- **Authentication**: Required (`BearerAuth` - Creator or payer only)
- **Description**: Partially update description, amount, currency, category, split type, or participant splits for an expense.
- **Path Parameters**:
  - `id` (string, required): Expense UUID.
- **Request Body**:
  ```json
  {
    "description": "Updated Dinner Title",
    "amount": 140.00,
    "currency": "USD",
    "category": "FOOD_AND_DRINK",
    "splitType": "EXACT",
    "splits": [
      { "userId": "usr-uuid-1", "amount": 80.00 },
      { "userId": "usr-uuid-2", "amount": 60.00 }
    ]
  }
  ```
- **Response** (`200 OK`): Updated `ExpenseWithSplits` object.
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden` (non-creator/payer), `404 Not Found`, `500 Internal Server Error`.

---

### 6. Delete Expense
- **DELETE** `/expenses/{id}`
- **Authentication**: Required (`BearerAuth` - Creator only)
- **Description**: Soft-delete an expense. Only the creator can perform this action.
- **Path Parameters**:
  - `id` (string, required): Expense UUID.
- **Response** (`204 No Content`).
- **Errors**: `401 Unauthorized`, `403 Forbidden` (non-creator), `404 Not Found`, `500 Internal Server Error`.

---

### 7. Get Balances & Simplified Debts
- **GET** `/balances?groupId={id}&simplified={bool}`
- **Authentication**: Required (`BearerAuth`)
- **Description**: Calculate net balances and recommended settlement transactions for a specific group or globally across all user groups and friendships.
- **Query Parameters**:
  - `groupId` (string, optional): Group UUID. If omitted, calculates global net balances across all groups and friends.
  - `simplified` (boolean, optional): Apply debt simplification algorithm (`true`/`false`). Default: `false`.
- **Response** (`200 OK`): `BalanceResponse` object.
  ```json
  {
    "balances": [
      {
        "userId": "usr-uuid-1",
        "userName": "Alice Vance",
        "netBalance": 60.00
      },
      {
        "userId": "usr-uuid-2",
        "userName": "Bob Smith",
        "netBalance": -60.00
      }
    ],
    "settlements": [
      {
        "fromUserId": "usr-uuid-2",
        "fromUserName": "Bob Smith",
        "toUserId": "usr-uuid-1",
        "toUserName": "Alice Vance",
        "amount": 60.00
      }
    ]
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `500 Internal Server Error`.
