# AuthGear
 
Work in progress

## Database setup

1. Setup dependencies:
   ```sh
   make vendor
   ```
2. Create a schema:
   ```sql
   CREATE SCHEMA app;
   ```
3. Setup environment variables (in `.env`):
   ```
   DATABASE_URL=postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable
   DATABASE_SCHEMA=app  
   ```
4. Apply database schema migrations:
   ```sh
   go run ./cmd/authgear migrate up
   ```
   
To create new migration:
```sh
# go run ./cmd/authgear migrate new <migration name>
go run ./cmd/authgear migrate new add user table
```