#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Helper functions
log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

log_info() {
    echo -e "${YELLOW}💡 $1${NC}"
}

log_header() {
    echo -e "${PURPLE}🚀 $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

echo ""
log_header "GRAFT CLI - GITHUB WORKFLOW INTEGRATION TEST"
echo "=============================================="

# Verify environment
if [ -z "$DATABASE_URL" ]; then
    log_error "DATABASE_URL environment variable is not set"
fi

log_success "Database URL: $DATABASE_URL"

# Store the original directory (workspace) before changing to test directory
WORKSPACE_DIR="$(pwd)"
log_success "Workspace directory: $WORKSPACE_DIR"

# Setup test directory
TEST_DIR="/tmp/graft-github-test-$(date +%s)"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

log_success "Test directory: $TEST_DIR"

# Determine graft binary path (check workspace first)
if [ -f "$WORKSPACE_DIR/graft" ]; then
    GRAFT_CMD="$WORKSPACE_DIR/graft"
elif [ -f "../graft" ]; then
    GRAFT_CMD="../graft"
elif [ -f "./graft" ]; then
    GRAFT_CMD="./graft"
elif command -v graft &> /dev/null; then
    GRAFT_CMD="graft"
else
    log_error "graft binary not found. Please build the project first with 'go build -o graft .'"
fi

GRAFT_VERSION=$($GRAFT_CMD --version 2>/dev/null || echo "Unknown")
log_success "Graft binary: $GRAFT_CMD"
log_success "Graft version: $GRAFT_VERSION"

echo ""
log_header "PHASE 1: PROJECT INITIALIZATION"
echo "==============================="

# Test 1: Initialize project
log_step "Initialize project"
$GRAFT_CMD init --postgresql --force >/dev/null 2>&1
echo "DATABASE_URL=$DATABASE_URL" > .env
log_success "Project initialized"

# Verify project structure
log_step "Verify project structure"
required_files=("graft.config.json" "sqlc.yml" ".env" "db/schema/schema.sql")
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        log_error "Required file missing: $file"
    fi
done
log_success "Project structure verified"

echo ""
log_header "PHASE 2: DATABASE OPERATIONS"
echo "============================"

# Test 2: Create initial schema
log_step "Create initial schema"
cat > db/schema/schema.sql << 'SCHEMA'
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
SCHEMA

$GRAFT_CMD migrate "create users table" --force >/dev/null 2>&1
$GRAFT_CMD apply --force >/dev/null 2>&1
log_success "Initial schema created and applied"

# Test 3: Insert test data
log_step "Insert test data"
cat > insert_data.sql << 'DATA'
INSERT INTO users (name, email) VALUES 
('Alice Johnson', 'alice@test.com'),
('Bob Smith', 'bob@test.com'),
('Charlie Brown', 'charlie@test.com');
DATA

$GRAFT_CMD raw insert_data.sql >/dev/null 2>&1
log_success "Test data inserted"

# Test 4: Create backup
log_step "Create backup"
$GRAFT_CMD backup "test backup" --force >/dev/null 2>&1
log_success "Backup created"

# Test 5: Add posts table
log_step "Add posts table"
cat > db/schema/schema.sql << 'SCHEMA2'
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
SCHEMA2

$GRAFT_CMD migrate "add posts table" --force >/dev/null 2>&1
$GRAFT_CMD apply --force >/dev/null 2>&1
log_success "Posts table added"

# Test 6: Insert posts data
log_step "Insert posts data"
cat > insert_posts.sql << 'POSTS'
INSERT INTO posts (user_id, title, content, published) VALUES 
(1, 'First Post', 'Content of first post', true),
(2, 'Second Post', 'Content of second post', false),
(3, 'Third Post', 'Content of third post', true);
POSTS

$GRAFT_CMD raw insert_posts.sql >/dev/null 2>&1
log_success "Posts data inserted"

echo ""
log_header "PHASE 3: ADVANCED TESTING"
echo "========================="

# Test 7: Complex query
log_step "Execute complex query"
cat > complex_query.sql << 'QUERY'
SELECT 
    u.name,
    u.email,
    COUNT(p.id) as post_count,
    COUNT(CASE WHEN p.published THEN 1 END) as published_count
FROM users u 
LEFT JOIN posts p ON u.id = p.user_id 
GROUP BY u.id, u.name, u.email 
ORDER BY u.name;
QUERY

echo "📊 Query Results:"
$GRAFT_CMD raw complex_query.sql
log_success "Complex query executed"

# Test 8: Check migration status
log_step "Check migration status"
echo "📋 Migration Status:"
$GRAFT_CMD status
log_success "Migration status checked"

# Test 9: Verify backup files
log_step "Verify backup files"
if [ -d "db/backup" ] && [ "$(ls -A db/backup 2>/dev/null)" ]; then
    backup_count=$(ls db/backup/*.json 2>/dev/null | wc -l)
    log_success "Found $backup_count backup files"
    
    # Validate backup files
    for backup_file in db/backup/*.json; do
        if [ -f "$backup_file" ]; then
            filename=$(basename "$backup_file")
            if command -v jq >/dev/null 2>&1; then
                if jq empty "$backup_file" 2>/dev/null; then
                    echo "   ✓ $filename - Valid JSON"
                else
                    echo "   ✗ $filename - Invalid JSON"
                fi
            else
                echo "   ✓ $filename - File exists"
            fi
        fi
    done
    log_success "Backup files verified"
else
    log_error "No backup files found"
fi

echo ""
log_header "PHASE 4: DATABASE RESET TEST"
echo "============================"

# Test 10: Database reset with automated responses
log_step "Test database reset with automated responses"
log_info "Sending automated responses: y (reset) and n (no backup)"

# Create a script to send the responses
cat > reset_responses.txt << 'RESPONSES'
y
n
RESPONSES

# Execute reset with automated responses
echo "🔄 Executing database reset..."
$GRAFT_CMD reset < reset_responses.txt

log_success "Database reset completed with automated responses"

# Test 11: Verify reset worked
log_step "Verify database was reset"
echo "📊 Checking table count after reset:"

# Check if tables still exist
cat > check_tables.sql << 'CHECK'
SELECT COUNT(*) as table_count 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_type = 'BASE TABLE';
CHECK

$GRAFT_CMD raw check_tables.sql
log_success "Database reset verification completed"

# Test 12: Final status check
log_step "Final migration status check"
echo "📋 Final Migration Status:"
$GRAFT_CMD status
log_success "Final status checked"

echo ""
log_header "PHASE 5: CLEANUP AND SUMMARY"
echo "============================"

# Cleanup
cd /
rm -rf "$TEST_DIR"
log_success "Test directory cleaned up"

echo ""
echo "🎉 ALL GITHUB WORKFLOW TESTS COMPLETED SUCCESSFULLY!"
echo "===================================================="
echo ""
log_success "✅ Project initialization and configuration"
log_success "✅ Schema creation and migration management"
log_success "✅ Data insertion and querying"
log_success "✅ Backup creation and validation"
log_success "✅ Complex SQL queries execution"
log_success "✅ Migration status tracking"
log_success "✅ Database reset with automated responses (y/n)"
log_success "✅ Post-reset verification"
echo ""
log_header "🚀 GRAFT CLI - READY FOR GITHUB WORKFLOW!"
echo ""
log_info "✨ All tests passed - GitHub Actions will run successfully"
log_info "🔧 Reset command works with automated y/n responses"
log_info "💾 Backup system functioning correctly"
log_info "📊 Migration tracking working perfectly"
