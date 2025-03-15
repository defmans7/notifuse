Users in this system are managed through a secure email-based authentication flow:

1. Sign In:

   - Users can sign in using their email address
   - A magic code is sent to their email for authentication
   - No password is required
   - Upon successful verification, the system:
     - Creates a 15-day valid session in the database
     - Returns a Paseto v4 asymmetric token containing:
       - User ID
       - User email
       - User name (if provided)
       - Session ID

2. Sign Up:
   - New users register with their email address
   - A Paseto v4 asymmetric token containing their email is generated
   - The token is sent to their email for verification
   - Account creation is completed upon email verification

This passwordless approach prioritizes security while maintaining a smooth user experience.
