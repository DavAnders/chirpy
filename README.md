# Chirpy API Documentation

Welcome to the Chirpy API! This document outlines the available endpoints within the Chirpy application and how to interact with them.

## Authentication

Many endpoints require authentication. This is achieved through a Bearer token provided in the `Authorization` header of your request.

## Endpoints

### User Management

- **Create User**
  - **POST** `/api/users`
  - Body: `{ "email": "user@example.com", "password": "password123" }`
  - Creates a new user with the provided email and password.

- **User Login**
  - **POST** `/api/login`
  - Body: `{ "email": "user@example.com", "password": "password123" }`
  - Authenticates the user and returns access and refresh JWT tokens.

- **Update User**
  - **PUT** `/api/users`
  - Headers: `Authorization: Bearer <access_token>`
  - Body: `{ "email": "newemail@example.com", "password": "newPassword123" }`
  - Updates the authenticated user's email and/or password.

### Chirp Management

- **Validate Chirp**
  - **POST** `/api/validate_chirp`
  - Body: `{ "body": "Chirp content" }`
  - Validates the chirp content without saving it.

- **Create Chirp**
  - **POST** `/api/chirps`
  - Headers: `Authorization: Bearer <access_token>`
  - Body: `{ "body": "Chirp content" }`
  - Creates a new chirp with the authenticated user as the author.

- **Get All Chirps**
  - **GET** `/api/chirps`
  - Optional Query: `?author_id=1`
  - Retrieves all chirps or those of a specific author if `author_id` is provided.

- **Get Chirp by ID**
  - **GET** `/api/chirps/{chirpID}`
  - Retrieves a specific chirp by its ID.

- **Delete Chirp**
  - **DELETE** `/api/chirps/{chirpID}`
  - Headers: `Authorization: Bearer <access_token>`
  - Deletes the chirp if the authenticated user is the author.

  - **Sorting Chirps** 
  - By default, chirps are sorted by their ID in ascending order (`asc`). You can modify the sorting order by adding a `sort` query parameter to the GET request, e.g., `/api/chirps?sort=desc` to sort chirps in descending order. Valid values for `sort` are `asc` for ascending order and `desc` for descending order.


### Token Management

- **Refresh Token**
  - **POST** `/api/refresh`
  - Headers: `Authorization: Bearer <refresh_token>`
  - Generates a new access token using a valid refresh token.

- **Revoke Token**
  - **POST** `/api/revoke`
  - Headers: `Authorization: Bearer <refresh_token>`
  - Revokes the provided refresh token.

### Webhooks

- **Polka Webhooks**
  - **POST** `/api/polka/webhooks`
  - Headers: `Authorization: ApiKey <Polka_API_Key>`
  - Handles Polka webhook events, specifically user upgrades.

### Admin Endpoints

- **Metrics**
  - **GET** `/admin/metrics`
  - Retrieves server metrics. Only accessible to admins.

- **Reset**
  - **HandleFunc** `/reset`
  - Resets the application state. Intended for development or testing.
