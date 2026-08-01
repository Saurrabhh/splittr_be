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
- **Response** (`201 Created`): `ExpenseWithSplits` — unified response shape `{ "expense": {...}, "splits": [...] }`.

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
- **Response** (`201 Created`): `ExpenseWithSplits` — same unified shape `{ "expense": {...}, "splits": [...] }`. For settlements, `splits` contains a single split.

### 3. List Expenses
- **GET** `/expenses?groupId={id}&personal={bool}&friendId={id}&limit={int}&cursor={string}`
- **Description**: Retrieve a cursor-paginated list of expenses with splits, filtered by group, personal budget, or direct friend. Splits for all items are bulk-fetched in a single query.
- **Response** (`200 OK`): `ListExpensesResponse` with `data` array of `ExpenseWithSplits` and pagination cursor.

### 4. Get Expense Details
- **GET** `/expenses/{id}`
- **Description**: Retrieve a specific expense record and its full split breakdown.
- **Response** (`200 OK`): `ExpenseWithSplits` — same unified shape `{ "expense": {...}, "splits": [...] }`.

### 5. Delete Expense
- **DELETE** `/expenses/{id}`
- **Description**: Soft-delete an expense. Only the creator of the expense can perform this action.
- **Response** (`204 No Content`).

### 6. Get Balances
- **GET** `/balances?groupId={id}&simplified={bool}`
- **Description**: Calculate net balances and recommended settlement transactions for a group or globally.
- **Response** (`200 OK`): `BalanceResponse`.
