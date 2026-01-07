# Backend Architecture Documentation

This document provides an overview of the Go backend service's architecture, its components, and how they interact.

## 1. Core Architecture: The Layered Approach

This project is built using a **Layered Architecture**, often referred to as Clean Architecture or Hexagonal Architecture. This design pattern separates the code into distinct layers, each with a specific responsibility.

Think of it like a skyscraper:

- **Handler (Lobby/Reception):** The public-facing entry point. It greets incoming HTTP requests, validates them, and directs them to the right department. It doesn't know how the work gets done, only who to ask.
- **Application Service (Manager's Office):** The business logic layer. It coordinates the work, enforces business rules, and makes decisions. It takes requests from the Handler and delegates the detailed tasks to the appropriate Repository.
- **Repository (Secure Vault):** The data access layer. Its only job is to get data from or put data into the database. It doesn't know or care about business rules; it just executes data operations.
- **Domain (The Blueprints):** This is the core of the application. It contains the fundamental data structures (`User` model) and the contracts (`UserRepository` interface) that the other layers must follow.

The most important rule is the **Dependency Rule**: Layers can only talk to the layer directly beneath them. A Handler talks to a Service, a Service talks to a Repository. This keeps the system organized, testable, and easy to modify.

## 2. Directory Structure

The project's code is organized within the `/internal` directory, which is standard for Go applications where the code isn't meant to be imported by other projects.

- **/cmd/server/main.go**: The main entry point of the application. Its responsibilities include:

  - Loading environment variables (`.env`).
  - Establishing a database connection.
  - **Dependency Injection (Wiring):** Creating instances of the repository, service, and handler, and passing them to each other.
  - Setting up the HTTP router (chi), including CORS and logging middleware.
  - Defining all API routes (e.g., `/api/login`, `/api/me`).
  - Starting the server.

- **/internal/domain**: The heart of the application.

  - **/model**: Defines the core data structures, like `user.go`. These are plain Go structs that represent our business entities.
  - **/repository**: Defines the interfaces that act as contracts for our data layer. For example, `user_repository.go` specifies that any user repository _must_ have methods like `FindByID` and `Save`.

- **/internal/application**: Contains the core business logic.

  - `user_service.go`: The `UserService` orchestrates user-related tasks like registration and login. It contains the business rules (e.g., checking if a password is correct) and calls the repository to fetch or save data.

- **/internal/repository/postgres**: An implementation of the repository interface defined in the domain layer.

  - `user_repo.go`: Contains the actual SQL queries to interact with the PostgreSQL database. If we ever wanted to switch to another database like MongoDB, we would simply create a new repository implementation without changing the service or handler layers.

- **/internal/handler**: Responsible for handling HTTP requests and responses.

  - `user_handler.go`: Contains the functions that are directly called by the router. They parse request bodies, call the appropriate `UserService` method, and format the HTTP response (e.g., sending back JSON or setting a cookie).
  - **/dto**: (Data Transfer Objects) Defines the structs for API request and response bodies, like `login_dto.go`. This separates our API contract from our internal domain models.

- **/internal/auth**: Handles JWT (JSON Web Token) creation and verification. It's a helper package used by the login service and authentication middleware.

- **/internal/middleware**: Contains HTTP middleware functions.
  - `jwt_middleware.go`: Protects routes by checking for a valid JWT in the request's cookie or header. It extracts the user ID and adds it to the request context for handlers to use.

## 3. Request Flow Example: User Login (`POST /api/login`)

1.  **Router (`main.go`):** An HTTP POST request hits `/api/login`. The `chi` router matches this and calls `userHandler.LoginUserHandler`.

2.  **Handler (`user_handler.go`):**

    - `LoginUserHandler` decodes the JSON request body into a `dto.LoginRequest` struct.
    - It calls the `userService.Login()` method, passing the request data.

3.  **Application Service (`login_service.go`):**

    - The `Login()` method receives the DTO.
    - It calls `userRepo.FindByUsernameOrEmail()` to fetch the user from the database.
    - It uses the `bcrypt` library to compare the stored password hash with the password from the request.
    - If the credentials are valid, it calls `auth.GenerateJWTToken()` to create a new token.
    - It returns the `user` model and the `token` string back to the handler.

4.  **Handler (`user_handler.go`):**
    - The handler receives the user and token from the service.
    - It sets the token in a secure, `HttpOnly` cookie on the HTTP response.
    - It creates a `dto.LoginResponse` struct (without the token) and sends it back to the client as a JSON response with a `200 OK` status.

This entire flow ensures that each part of the system only does its own job, making the code clean, secure, and easy to debug.

## 4. Authentication and Authorization

Authentication (proving who you are) and Authorization (checking what you're allowed to do) are handled via JSON Web Tokens (JWT) and custom middleware.

### JWT Generation (`/internal/auth/jwt.go`)

- **`GenerateJWTToken`**: This function is called by the `LoginService` after a user's credentials have been successfully verified.
  - It creates a new JWT containing several "claims" or pieces of information:
    - `ID`: The user's unique ID from the database.
    - `Username`: The user's username.
    - `Role`: The user's role (e.g., 'admin', 'user').
    - `ExpiresAt`: A future timestamp (e.g., 15 minutes from now) after which the token is no longer valid.
  - The entire token is then digitally signed using a secret key (`JWT_SIGNING_KEY` from your `.env` file). This signature prevents the token from being tampered with.
  - The final, signed token string is returned.

### JWT Verification (`/internal/auth/jwt.go`)

- **`VerifyJWTToken`**: This function is used by our middleware to validate an incoming token.
  - It parses the token string.
  - It checks the token's signature using the same secret key to ensure it hasn't been modified.
  - It verifies that the token has not expired.
  - If the token is valid, it returns the claims (ID, Username, Role) for the middleware to use.

### Authentication Middleware (`/internal/middleware/jwt_middleware.go`)

- **`JWTAuthMiddleware`**: This is a `chi` middleware that protects specific routes (like `/api/me`). It acts as a gatekeeper for every incoming request to a protected endpoint.
  1.  **Token Extraction**: It first tries to find the JWT from a secure, `HttpOnly` cookie named `access_token`. If not found, it checks for a standard `Authorization: Bearer <token>` header.
  2.  **Validation**: If a token is found, it calls `auth.VerifyJWTToken()` to validate it. If the token is missing, invalid, or expired, the middleware immediately stops the request and sends a `401 Unauthorized` error.
  3.  **Context Injection**: If the token is valid, the middleware extracts the user's ID and role from the token's claims. It then adds this information to the request's `context`.
  4.  **Passing the Request**: Finally, it calls `next.ServeHTTP()` to pass the request along to the next handler in the chain (e.g., `MeHandler`).

### Authorization (Role-Based Access)

- **`RoleMiddleware`**: This middleware would be used to protect routes that require a specific user role (e.g., an admin dashboard).
  - It would run _after_ `JWTAuthMiddleware`.
  - It would extract the user's role from the request context (put there by the `JWTAuthMiddleware`).
  - It would check if the user's role is in the list of allowed roles for that endpoint.
  - If the role is not allowed, it would stop the request and send a `403 Forbidden` error.

This combination of JWTs and middleware provides a robust and secure way to manage access to your API.
