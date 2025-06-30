# Roles API Documentation

## Base Endpoint

**`/api/v2/namespaces/{namespace}/roles`**

This API allows for managing roles within a specific namespace. It supports creating, retrieving, listing, updating, and deleting roles.

---

## Endpoints

### 1. Create a New Role

**POST** `/api/v2/namespaces/{namespace}/roles`

#### Request Body:
```json
{
  "name": "foo1",
  "description": "foo1 description",
  "oidcGroups": ["foo1_g1", "foo1_g2"],
  "permissions": [
    {
      "topic": "secrets",
      "method": "read"
    },
    {
      "topic": "variables",
      "method": "manage"
    }
  ]
}
```

#### Response:
**Status Code:** `200 OK`
```json
{
  "data": {
    "name": "foo1",
    "description": "foo1 description",
    "oidcGroups": ["foo1_g1", "foo1_g2"],
    "permissions": [
      {
        "topic": "secrets",
        "method": "read"
      },
      {
        "topic": "variables",
        "method": "manage"
      }
    ],
    "createdAt": "2024-02-05T12:00:00Z",
    "updatedAt": "2024-02-05T12:00:00Z"
  }
}
```

---

### 2. Get a Role

**GET** `/api/v2/namespaces/{namespace}/roles/{roleName}`

#### Response:
**Status Code:** `200 OK`
```json
{
  "data": {
    "name": "foo1",
    "description": "foo1 description",
    "oidcGroups": ["foo1_g1", "foo1_g2"],
    "permissions": [
      {
        "topic": "secrets",
        "method": "read"
      },
      {
        "topic": "variables",
        "method": "manage"
      }
    ],
    "createdAt": "2024-02-05T12:00:00Z",
    "updatedAt": "2024-02-05T12:00:00Z"
  }
}
```

---

### 3. List All Roles

**GET** `/api/v2/namespaces/{namespace}/roles`

#### Response:
**Status Code:** `200 OK`
```json
{
  "data": [
    {
      "name": "foo1",
      "description": "foo1 description",
      "oidcGroups": ["foo1_g1", "foo1_g2"],
      "permissions": [
        {
          "topic": "secrets",
          "method": "read"
        },
        {
          "topic": "variables",
          "method": "manage"
        }
      ],
      "createdAt": "2024-02-05T12:00:00Z",
      "updatedAt": "2024-02-05T12:00:00Z"
    },
    {
      "name": "foo2",
      "description": "foo2 description",
      "oidcGroups": ["foo2_g1", "foo2_g2"],
      "permissions": [
        {
          "topic": "secrets",
          "method": "read"
        },
        {
          "topic": "variables",
          "method": "manage"
        }
      ],
      "createdAt": "2024-02-05T12:00:00Z",
      "updatedAt": "2024-02-05T12:00:00Z"
    }
  ]
}
```

---

### 4. Update a Role

**PUT** `/api/v2/namespaces/{namespace}/roles/{roleName}`

#### Request Body:
```json
{
  "name": "foo3",
  "description": "Updated description",
  "oidcGroups": ["foo3_g1", "foo3_g2"],
  "permissions": [
    {
      "topic": "secrets",
      "method": "read"
    },
    {
      "topic": "variables",
      "method": "manage"
    }
  ]
}
```

#### Response:
**Status Code:** `200 OK`
```json
{
  "data": {
    "name": "foo3",
    "description": "Updated description",
    "oidcGroups": ["foo3_g1", "foo3_g2"],
    "permissions": [
      {
        "topic": "secrets",
        "method": "read"
      },
      {
        "topic": "variables",
        "method": "manage"
      }
    ],
    "createdAt": "2024-02-05T12:00:00Z",
    "updatedAt": "2024-02-05T12:00:00Z"
  }
}
```

---

### 5. Delete a Role

**DELETE** `/api/v2/namespaces/{namespace}/roles/{roleName}`

#### Response:
**Status Code:** `200 OK`
```json
{}
```

---

## Notes:
- Role names should be unique within a namespace.
- Field `method` should be either "read" or "manage". 
---

## Example Usage:
### Creating a Role via `curl`
```sh
curl -X POST "http://localhost/api/v2/namespaces/test/roles" \
     -H "Content-Type: application/json" \
     -d '{"name": "foo1", "description": "A test role", "oidcGroups": ["group1"], "permissions": [{"topic": "test", "method": "read"}]}'
```

