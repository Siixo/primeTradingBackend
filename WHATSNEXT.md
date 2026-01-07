### WHAT'S NEXT ?

##Frontend dev
Register sends an error:
duplicate key value violates unique constraint "user_pkey"

Login sends an error:
Error during login: SyntaxError: JSON.parse: unexpected character at line 1 column 1 of the JSON data

##WIP:
#Refresh token:
Refresh token func written
Need to implement service and handler
Check incoming credentials (claims) and token
Validate them, then renew token

Token refresh validation process:
How the Cookie is Set Up
When a user successfully logs in through your Login handler, the following process takes place:

**Token
Generation:** Your server first generates a JSON Web Token (JWT). This token is a secure, digitally signed string that contains information about the user (like their user ID) and an expiration time.

Cookie Creation: The server then creates an HTTP cookie named access_token. The JWT generated in the previous step is set as the value of this cookie.

Cookie Configuration: This cookie is configured with several important attributes for security and functionality:

HttpOnly: This is a crucial security setting. It prevents any JavaScript running in the browser from accessing the cookie. This is a strong defense against Cross-Site Scripting (XSS) attacks where an attacker might try to steal the token.
Path: The cookie's path is set to /, which means the browser will send this cookie with every request made to your server, regardless of the specific page or endpoint.
Expires/Max-Age: The cookie is given a limited lifespan (for example, one hour). After this time, the browser automatically deletes the cookie, and the user will need to log in again.
The browser stores this cookie and automatically includes it in the headers of all subsequent requests to your server.

How the Token is Extracted
When the browser sends a request to a protected endpoint on your server, a middleware function (specifically, your jwt_middleware.go) runs before the main request handler.

Cookie Retrieval: The middleware inspects the incoming request for the cookie named access_token.

Token Extraction: If the cookie is found, the middleware extracts its value, which is the JWT string. Your setup also includes a fallback to check the Authorization header for a "Bearer" token, which is a common alternative way to send tokens.

Token Validation: Once the token string is extracted, the middleware validates it by:

Checking the token's signature with a secret key known only to the server. This verifies that the token is authentic and has not been modified.
Checking the expiration time to ensure the token is still valid.
If the token is valid, the middleware extracts the user information from it and passes the request along to the intended handler. If the cookie is missing or the token is invalid, the middleware rejects the request, typically with an "Unauthorized" error.
