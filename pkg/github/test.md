```graphql
mutation changeUserStatus {
  changeUserStatus(input: $status) {
    status {
      id
      updatedAt
      expiresAt
    }
  }
}

mutation changeUserStatus {
    changeUserStatus($status: ChangeUserStatusInput!) {
        status {
            id
            updatedAt
            expiresAt
        }
    }
}

```


```graphql
{
  "status": {
    "clientMutationId": "test",
   "emoji": ":green_circle:",
  "expiresAt": "2023-05-10T10:15:30Z",
  "limitedAvailability": false,
    "message": "test",
  "organizationId": ""
  }
}
```
