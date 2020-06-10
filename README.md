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
   ./migrate/migrate up
   ```
   
To create new migration:
```sh
# ./migrate/migrate new <migration name>
./migrate/migrate new add user table
```