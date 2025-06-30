# API Documentation: /api_tokens

## Overview
The `/api_tokens` API provides endpoints to manage API tokens within a namespace. It supports operations to create, retrieve, list, and delete API tokens.

## Base URL
`/api/v2/namespaces/{namespace}/api_tokens`

## Endpoints

### 1. Create a New API Token
**Endpoint:**
```
POST /api/v2/namespaces/{namespace}/api_tokens
```

**Request Body:**
```json
{
  "name": "foo1",
  "description": "foo1 description",
  "permissions": [
    {"topic": "foo1_topic1", "method": "foo1_method1"},
    {"topic": "foo1_topic2", "method": "foo1_method2"}
  ],
  "duration": "P1DT2H30M"
}
```

**Response:**
```json
{
  "data": {
    "apiToken": {
      "name": "foo1",
      "description": "foo1 description",
      "prefix": "generated_prefix",
      "permissions": [
        {"topic": "foo1_topic1", "method": "foo1_method1"},
        {"topic": "foo1_topic2", "method": "foo1_method2"}
      ],
      "expiredAt": "timestamp",      
      "isExpired": "boolean",
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    },
    "secret": "generated_secret"
  }
}
```
---

### 2. Retrieve an API Token
**Endpoint:**
```
GET /api/v2/namespaces/{namespace}/api_tokens/{token_name}
```

**Response:**
```json
{
  "data": {
    "name": "foo1",
    "description": "foo1 description",
    "prefix": "generated_prefix",
    "permissions": [
      {"topic": "foo1_topic1", "method": "foo1_method1"},
      {"topic": "foo1_topic2", "method": "foo1_method2"}
    ],
    "expiredAt": "timestamp",      
    "isExpired": "boolean",
    "createdAt": "timestamp",
    "updatedAt": "timestamp"
  }
}
```
---

### 3. List All API Tokens
**Endpoint:**
```
GET /api/v2/namespaces/{namespace}/api_tokens
```

**Response:**
```json
{
  "data": [
    {
      "name": "foo1",
      "description": "foo1 description",
      "prefix": "generated_prefix",
      "permissions": [
        {"topic": "foo1_topic1", "method": "foo1_method1"},
        {"topic": "foo1_topic2", "method": "foo1_method2"}
      ],
      "expiredAt": "timestamp",      
      "isExpired": "boolean",
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    },
    {
      "name": "foo2",
      "description": "foo2 description",
      "prefix": "generated_prefix",
      "permissions": [
        {"topic": "foo2_topic1", "method": "foo2_method1"},
        {"topic": "foo2_topic2", "method": "foo2_method2"}
      ],
      "expiredAt": "timestamp",      
      "isExpired": "boolean",
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    }
  ]
}
```
---

### 4. Delete an API Token
**Endpoint:**
```
DELETE /api/v2/namespaces/{namespace}/api_tokens/{token_name}
```

## Notes
- The `secret` returned when creating an API token is only shown once and should be stored securely.
- API tokens are tied to a namespace and cannot be accessed outside their assigned namespace.
- Each API token includes `permissions`, defining the topics and methods it can access.
- field `duration` in the post request should be in ISO8601 format.

