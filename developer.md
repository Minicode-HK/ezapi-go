Collecting workspace information# Functions in app.go

Here are all the functions exported from app.go:

## Routes Setup Functions

- **`New[T any](db *[]T) *App[T]`** - Creates a new ez instance for a module
- **`(a *App[T]) Seed(initialData []T) *App[T]`** - Seeds the database with initial data. If /api/reset is called, it will reset to this data.
- **`(a *App[T]) CRUD(path string) *App[T]`** - Registers CRUD endpoints for the resource
  - the returned data will be have this type of format: `{ "data": [...] , "success": ... }`
- **`(a *App[T]) CustomRoutes(fn func(*gin.Engine)) *App[T]`** - Adds custom route handlers


## Router Application Function
- **`SetupAllRouters(router *gin.Engine)`** - Applies all registered routers to the engine. Call this in `main.go`.

## Helper Functions (HTTP Response)

- **`SendSuccess(c *gin.Context, data any)`** - Sends a successful JSON response
  - **Response Format:**
    ```json
    {
      "success": true,
      "data": <your_data>,
      "message": "Success"
    }
    ```

- **`SendError(c *gin.Context, statusCode int, message string)`** - Sends an error JSON response
  - **Response Format:**
    ```json
    {
      "success": false,
      "error": "<error_message>",
      "statusCode": <http_status_code>
    }
    ```

- **`SendErrorWithDetails(c *gin.Context, statusCode int, message string, details any)`** - Sends an error response with detailed information
  - **Response Format:**
    ```json
    {
      "success": false,
      "error": "<error_message>",
      "statusCode": <http_status_code>,
      "details": <detailed_error_info>
    }
    ```
## Query Builder

- **`Query[T any](data []T) *feature.QueryBuilder[T]`** - Creates a query builder for filtering/sorting data

### QueryBuilder Methods

- **`(qb *QueryBuilder[T]) Get() []T`** - Returns the current filtered/sorted results as a slice

- **`(qb *QueryBuilder[T]) First() *T`** - Returns a pointer to the first item in the result set, or nil if empty

- **`(qb *QueryBuilder[T]) Filter(filterFunc func(*T) bool) *QueryBuilder[T]`** - Filters data using a custom function that receives a pointer to each item
  - The function should return `true` to include the item, `false` to exclude it

- **`(qb *QueryBuilder[T]) Where(fieldName string, operator string, value any) *QueryBuilder[T]`** - Filters data by field name, operator, and value
  - Supported operators: `=`, `==`, `!=`, `>`, `<`, `>=`, `<=`

- **`(qb *QueryBuilder[T]) WhereWith(predicate func(*T) bool) *QueryBuilder[T]`** - Alias for `Filter()`. Filters using a custom predicate function

- **`(qb *QueryBuilder[T]) OrderBy(fieldName string, ascending ...bool) *QueryBuilder[T]`** - Sorts results by field name
  - Second parameter (optional): `true` for ascending (default), `false` for descending

- **`(qb *QueryBuilder[T]) Limit(n int) *QueryBuilder[T]`** - Limits the number of results to `n`. Directly affects pagination.


- **`(qb *QueryBuilder[T]) Offset(n int) *QueryBuilder[T]`** - Skips the first `n` results (pagination)

- **`(qb *QueryBuilder[T]) Count() int`** - Returns the total count of items in the current result set

- **`(qb *QueryBuilder[T]) Select(fields ...string) *QueryBuilder[map[string]any]`** - Projects selected fields into a new result set as maps
  - Returns a `QueryBuilder[map[string]any]` with only the specified fields
  - Respects JSON struct tags if present

## Hook Functions (Lifecycle Callbacks)

- Note: These 'After' hooks WILL NOT modify the data in the database, they are for side effects only.

- **`(a *App[T]) BeforeCreate(hook feature.BeforeCreateFunc[T]) *App[T]`** - Runs before creating an item
- **`(a *App[T]) AfterCreate(hook feature.AfterCreateFunc[T]) *App[T]`** - Runs after creating an item
- **`(a *App[T]) BeforeUpdate(hook feature.BeforeUpdateFunc[T]) *App[T]`** - Runs before updating an item
- **`(a *App[T]) AfterUpdate(hook feature.AfterUpdateFunc[T]) *App[T]`** - Runs after updating an item
- **`(a *App[T]) BeforeDelete(hook feature.BeforeDeleteFunc[T]) *App[T]`** - Runs before deleting an item
- **`(a *App[T]) AfterDelete(hook feature.AfterDeleteFunc[T]) *App[T]`** - Runs after deleting an item

## Authentication Functions

- **`SetupAuthProvider`** - Configures JWT authentication provider. Call this in `main.go`.
- **`SetUserDB(users []auth.User)`** - Sets the user database for authentication.`

## CMS Functions

- **`RegisterCMSBackend`** - Registers CMS backend API routes
- **`RegisterCMSFrontend`** - Registers CMS frontend UI routes
- **`RegisterCMS(router *gin.Engine)`** - Registers both CMS backend and frontend

## Utility Functions

- **`(ez *App[T]) AutoTimestamp() *App[T]`** - Automatically sets `CreatedAt` and `UpdatedAt` timestamps on create/update. You need to have `CreatedAt` and `UpdatedAt` fields in your struct for this to work.