# Expense API Documentation

The Expense module manages shared group and individual expenses, splitting logic (Equal, Exact, Percentage), balance calculations, debt simplification, and settlement payments.

## Endpoints

### 1. Create Expense
- **POST** `/expenses`
- **Description**: Log a new expense and distribute the splits across users.
- **Request Body**:
  ```json
  {
    "description": "Dinner at Mario's",
    "amount": 120.00,
    "currency": "USD",
    "category": "Food",
    "groupId": "grp_123",
    "paidBy": "usr_abc",
    "splitType": "EQUAL",
    "splits": [
      { "userId": "usr_abc" },
      { "userId": "usr_xyz" }
    ]
  }
  ```
- **Response** (`201 Created`): `ExpenseWithSplits` — unified response shape where expense fields are at root along with `splits`: `{ "id": "...", "description": "...", "amount": 120.00, ..., "splits": [...] }`.

### 2. Settle Up
- **POST** `/expenses/settle`
- **Description**: Record a payment transaction between two users to clear or reduce debt.
- **Request Body**:
  ```json
  {
    "amount": 50.00,
    "currency": "USD",
    "groupId": "grp_123",
    "paidBy": "usr_xyz",
    "receivedBy": "usr_abc"
  }
  ```
- **Response** (`201 Created`): `ExpenseWithSplits` — flat response shape `{ "id": "...", "amount": 50.00, ..., "splits": [...] }`. For settlements, `splits` contains a single split.

### 3. List Expenses
- **GET** `/expenses?groupId={id}&personal={bool}&friendId={id}&limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of expenses with splits, filtered by group, personal budget, or direct friend. Splits for all items are bulk-fetched in a single query. Filter parameters (`groupId`, `personal`, `friendId`) are mutually exclusive — passing more than one returns `400 Bad Request`.
- **Response** (`200 OK`): paginated `Page` envelope — `data` array of `ExpenseWithSplits` plus `pagination` metadata (`nextCursor`, `hasMore`).


### 4. Get Expense Details
- **GET** `/expenses/{id}`
- **Description**: Retrieve a specific expense record and its full split breakdown.
- **Response** (`200 OK`): `ExpenseWithSplits` — flat response shape `{ "id": "...", "amount": ..., ..., "splits": [...] }`.

### 5. Delete Expense
- **DELETE** `/expenses/{id}`
- **Description**: Soft-delete an expense. Only the creator of the expense can perform this action.
- **Response** (`204 No Content`).

### 6. Get Balances
- **GET** `/balances?groupId={id}&simplified={bool}`
- **Description**: Calculate net balances and recommended settlement transactions for a group or globally.
- **Response** (`200 OK`): `BalanceResponse`.

### 7. Sync Expenses
- **GET** `/expenses/sync?lastVersion={int64}&limit={int}`
- **Description**: Retrieve active and soft-deleted expenses modified after a given sequence version for offline sync.
- **Response** (`200 OK`): `ExpenseSyncResponse` (`{ "newVersion": 150, "updated": [...], "deletedIds": [...] }`).

### 8. Update Expense
- **PATCH** `/expenses/{id}`
- **Description**: Partially update description, amount, currency, category, or splits for an expense. Allowed only for the user who created or paid for the expense.
- **Request Body**:
  ```json
  {
    "description": "Updated Dinner Title",
    "amount": 135.00,
    "splits": [
      { "userId": "usr_abc" },
      { "userId": "usr_xyz" }
    ]
  }
  ```
- **Response** (`200 OK`): `ExpenseWithSplits` object.

