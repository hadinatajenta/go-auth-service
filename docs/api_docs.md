# Auth Service API Documentation

This document provides a summary of all available API endpoints in the `auth-service`.

## Base URL

`http://localhost:8080/api/v1`

## Health Check

- **URL**: `/health`
- **Method**: `GET`
- **Auth**: None
- **Success Response**: `{"message": "Server is healthy", "data": {"status": "up"}}`

---

## Authentication

### Login

- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth**: None
- **Body**:
  ```json
  {
    "username": "your_username",
    "password": "your_password"
  }
  ```

### Register

- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth**: None
- **Body**:
  ```json
  {
    "username": "new_username",
    "email": "user@example.com",
    "password": "strong_password",
    "first_name": "John",
    "last_name": "Doe"
  }
  ```

---

## Users (Protected)

_Requires Bearer Token_

- **GET /me**: Get current user profile.
- **GET /users**: List all users.
- **GET /users/:id**: Get user profile by ID.
- **PUT /users/:id**: Update user profile (First Name, Last Name).
- **DELETE /users/:id**: Delete user.

---

## Roles (Protected)

_Requires Bearer Token_

- **POST /roles**: Create a new role.
- **GET /roles**: List all roles.
- **GET /roles/:id**: Get role by ID.
- **PUT /roles/:id**: Update role.
- **DELETE /roles/:id**: Delete role.

---

## Permissions (Protected)

_Requires Bearer Token_

- **POST /permissions**: Create a new permission.
- **GET /permissions**: List all permissions.
- **GET /permissions/:id**: Get permission by ID.
- **PUT /permissions/:id**: Update permission.
- **DELETE /permissions/:id**: Delete permission.

---

## Menus (Protected)

_Requires Bearer Token_

- **POST /menus**: Create a new menu item.
- **GET /menus**: List all menu items.
- **GET /menus/:id**: Get menu by ID.
- **PUT /menus/:id**: Update menu item.
- **DELETE /menus/:id**: Delete menu item.
