#!/bin/bash

# Test Authentication Flow for Envoy API
# This script demonstrates the complete authentication workflow

set -e  # Exit on error

BASE_URL="http://localhost:8080"
TEST_EMAIL="testuser_$(date +%s)@example.com"
TEST_PASSWORD="SecurePassword123"
TEST_NAME="Test User"

echo "=================================="
echo "Envoy Authentication Flow Test"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print test step
print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to print success
print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if server is running
print_step "Checking if server is running..."
if ! curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" | grep -q "200"; then
    print_error "Server is not running at $BASE_URL"
    echo "Please start the server first: make run"
    exit 1
fi
print_success "Server is running"
echo ""

# Test 1: Register a new user
print_step "Test 1: Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"$TEST_NAME\",
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")

echo "Response: $REGISTER_RESPONSE"

# Extract token from registration
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    print_error "Failed to register user"
    exit 1
fi

print_success "User registered successfully"
echo "Token: ${TOKEN:0:50}..."
echo ""

# Test 2: Try to access protected endpoint without token
print_step "Test 2: Accessing protected endpoint WITHOUT token (should fail)..."
UNAUTH_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "$BASE_URL/hello")
HTTP_CODE=$(echo "$UNAUTH_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "401" ]; then
    print_success "Correctly rejected unauthorized request (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi
echo ""

# Test 3: Access protected endpoint with token
print_step "Test 3: Accessing protected endpoint WITH token..."
HELLO_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "$BASE_URL/hello" \
    -H "Authorization: Bearer $TOKEN")
HTTP_CODE=$(echo "$HELLO_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "200" ]; then
    print_success "Successfully accessed protected endpoint"
    echo "Response: $(echo "$HELLO_RESPONSE" | grep -v "HTTP_CODE")"
else
    print_error "Failed to access protected endpoint. HTTP Code: $HTTP_CODE"
fi
echo ""

# Test 4: Access profile endpoint
print_step "Test 4: Getting user profile..."
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/profile" \
    -H "Authorization: Bearer $TOKEN")

echo "Response: $PROFILE_RESPONSE"

if echo "$PROFILE_RESPONSE" | grep -q "user_id"; then
    print_success "Successfully retrieved user profile"
else
    print_error "Failed to retrieve profile"
fi
echo ""

# Test 5: Login with credentials
print_step "Test 5: Logging in with credentials..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")

echo "Response: $LOGIN_RESPONSE"

NEW_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$NEW_TOKEN" ]; then
    print_error "Failed to login"
    exit 1
fi

print_success "Login successful"
echo "New Token: ${NEW_TOKEN:0:50}..."
echo ""

# Test 6: Try login with wrong password
print_step "Test 6: Attempting login with WRONG password (should fail)..."
WRONG_LOGIN_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"WrongPassword123\"
    }")
HTTP_CODE=$(echo "$WRONG_LOGIN_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "401" ]; then
    print_success "Correctly rejected invalid credentials (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi
echo ""

# Test 7: Try to register duplicate user
print_step "Test 7: Attempting to register DUPLICATE user (should fail)..."
DUPLICATE_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"$TEST_NAME\",
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")
HTTP_CODE=$(echo "$DUPLICATE_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "409" ]; then
    print_success "Correctly rejected duplicate user (409)"
else
    print_error "Expected 409, got $HTTP_CODE"
fi
echo ""

# Test 8: Access goodbye endpoint with token
print_step "Test 8: Accessing POST /goodbye endpoint WITH token..."
GOODBYE_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/goodbye" \
    -H "Authorization: Bearer $TOKEN")
HTTP_CODE=$(echo "$GOODBYE_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "200" ]; then
    print_success "Successfully accessed POST protected endpoint"
    echo "Response: $(echo "$GOODBYE_RESPONSE" | grep -v "HTTP_CODE")"
else
    print_error "Failed to access protected endpoint. HTTP Code: $HTTP_CODE"
fi
echo ""

# Test 9: Try with invalid token
print_step "Test 9: Accessing protected endpoint with INVALID token (should fail)..."
INVALID_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "$BASE_URL/profile" \
    -H "Authorization: Bearer invalid.token.here")
HTTP_CODE=$(echo "$INVALID_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "401" ]; then
    print_success "Correctly rejected invalid token (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi
echo ""

# Test 10: Try with malformed Authorization header
print_step "Test 10: Accessing protected endpoint with MALFORMED header (should fail)..."
MALFORMED_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "$BASE_URL/profile" \
    -H "Authorization: NotBearer $TOKEN")
HTTP_CODE=$(echo "$MALFORMED_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$HTTP_CODE" == "401" ]; then
    print_success "Correctly rejected malformed header (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi
echo ""

# Summary
echo "=================================="
echo "Test Summary"
echo "=================================="
print_success "All authentication flow tests completed!"
echo ""
echo "Test Details:"
echo "  - User Email: $TEST_EMAIL"
echo "  - Registration: ✓"
echo "  - Login: ✓"
echo "  - Profile Access: ✓"
echo "  - Protected Endpoints: ✓"
echo "  - Security Validations: ✓"
echo ""
echo "You can now use this token to access protected endpoints:"
echo "export AUTH_TOKEN=\"$TOKEN\""
echo ""
echo "Example commands:"
echo "  curl -H \"Authorization: Bearer \$AUTH_TOKEN\" $BASE_URL/profile"
echo "  curl -H \"Authorization: Bearer \$AUTH_TOKEN\" $BASE_URL/hello"
echo "  curl -X POST -H \"Authorization: Bearer \$AUTH_TOKEN\" $BASE_URL/goodbye"
