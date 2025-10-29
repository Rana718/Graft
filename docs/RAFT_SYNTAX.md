# Graft Schema Syntax (.raft)

A beginner-friendly, Prisma-like schema definition language for Graft.

## Basic Syntax

```raft
model ModelName = {
  fieldName FieldType @attribute
}
```

## Field Types

| Raft Type | PostgreSQL | MySQL | SQLite |
|-----------|------------|-------|--------|
| `String` | VARCHAR(255) | VARCHAR(255) | TEXT |
| `Text` | TEXT | TEXT | TEXT |
| `Int` | INTEGER | INT | INTEGER |
| `BigInt` | BIGINT | BIGINT | INTEGER |
| `Float` | FLOAT | FLOAT | REAL |
| `Decimal` | DECIMAL(10,2) | DECIMAL(10,2) | REAL |
| `Boolean` | BOOLEAN | TINYINT(1) | INTEGER |
| `DateTime` | TIMESTAMP | DATETIME | TEXT |
| `Date` | DATE | DATE | TEXT |
| `Json` | JSONB | JSON | TEXT |

## Attributes

### `@id`
Marks field as primary key with auto-increment

```raft
model User = {
  id Int @id
}
```

### `@unique`
Ensures field values are unique

```raft
model User = {
  email String @unique
}
```

### `@default`
Sets default value for field

```raft
model User = {
  createdAt DateTime @default
  isActive Boolean @default
  role String @default("user")
}
```

### `@relation(Table.column)`
Creates foreign key relationship

```raft
model Post = {
  authorId Int @relation(User.id)
}
```

### `@optional`
Allows NULL values (fields are NOT NULL by default)

```raft
model User = {
  bio Text @optional
}
```

## Complete Example

```raft
// User model
model User = {
  id Int @id
  email String @unique
  name String
  password String
  bio Text @optional
  role String @default("user")
  isActive Boolean @default
  createdAt DateTime @default
  updatedAt DateTime @default
}

// Post model with relations
model Post = {
  id Int @id
  title String
  content Text
  slug String @unique
  published Boolean @default
  views Int @default(0)
  authorId Int @relation(User.id)
  createdAt DateTime @default
  updatedAt DateTime @default
}

// Comment model
model Comment = {
  id Int @id
  content Text
  postId Int @relation(Post.id)
  userId Int @relation(User.id)
  createdAt DateTime @default
}
```

## Workflow

1. **Create schema.raft**
```bash
# Create your schema file
touch schema.raft
```

2. **Define your models**
```raft
model User = {
  id Int @id
  email String @unique
  name String
}
```

3. **Generate SQL**
```bash
graft generate
```

4. **Create migration**
```bash
graft migrate "initial schema"
```

5. **Apply migration**
```bash
graft apply
```

## Comments

Use `//` for single-line comments:

```raft
// This is a comment
model User = {
  id Int @id  // Primary key
  email String @unique  // Must be unique
}
```

## Best Practices

1. Always use `@id` for primary keys
2. Use `@unique` for fields like email, username
3. Use `@relation` for foreign keys
4. Use `@default` for timestamps and boolean flags
5. Use `@optional` sparingly - prefer NOT NULL when possible
6. Use descriptive model and field names
7. Keep related models together in the file

## Migration from SQL

If you have existing SQL schemas, you can:

1. Use `graft pull` to extract current schema
2. Manually convert SQL to `.raft` syntax
3. Use `graft generate` to create new SQL
4. Compare and create migration for differences
