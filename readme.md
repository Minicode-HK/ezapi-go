# EZAPI-GO 

![ezapi-go-logo](./docs/landing.png)

> **Build and test REST APIs in minutes, not hours.**  
> Zero configuration. Zero database setup. Just pure Go magic.

```bash
git clone https://github.com/Minicode-HK/ezapi-go.git
cd ezapi-go
go get
go run main.go
# Your API is live at http://localhost:8080
```

---

## The Problem

**Traditional API Development:**
```
❌ Install & configure database     
❌ Write migrations                 
❌ Setup ORM/SQL queries            
❌ Build CRUD endpoints             
❌ Create admin interface           
❌ Write API documentation          
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Takes hours to days. Frustrating!
```

**With EZAPI-GO:**
```
✅ Define your struct                (2 min)
✅ Add one line: RegisterRouter()   (10 sec)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Total: 2 minutes. API + Admin ready.
```

---

## Quick Start

### 1. Define Your Model
```go
// route/product.go
type Product struct {
    Id       string  `json:"id"`
    Name     string  `json:"name" binding:"required"`
    Price    float64 `json:"price" binding:"required,gt=0"`
    Category string  `json:"category"`
}

var ProductDB []Product
```

### 2. Register (One Line)
```go
func init() {
    ProductDB = ResetableDatabase(&ProductDB, []Product{
        {Id: "1", Name: "Laptop", Price: 999.99, Category: "Electronics"},
    })
    
    RegisterRouter(&ProductDB, "/api/products") // That's it!
}
```

### 3. You're Done!
```bash
# 5 REST endpoints created automatically:
GET    /api/products       # List all
GET    /api/products/:id   # Get by ID
POST   /api/products       # Create
PUT    /api/products/:id   # Update
DELETE /api/products/:id   # Delete

# Plus: Full admin UI at /cms
```

---

## Perfect For Anyone

<table>
<tr>
<td width="33%">

### Frontend Devs
```go
// Mock API in 30 seconds
type User struct {
    Id    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
RegisterRouter(&UserDB, "/api/users")
```
**Focus on React/Vue, not backend**

</td>

<td width="33%">

### Proof of Concepts
```go
// Try ideas fast
type Post struct { ... }
type Comment struct { ... }

RegisterRouter(&PostDB, "/api/posts")
RegisterRouter(&CommentDB, "/api/comments")
```
**Validate concepts before building**

</td>
</tr>
</table>

---

## Wonderful Admin Panel

**Every model gets a full-featured CMS automatically:**

![Admin Interface](./docs/create_form.png)

### ✨ Features

<table>
<tr>
<td width="50%">

#### Data Management
- **Auto CRUD UI** - List, Create, Edit, Delete
- **Smart Forms** - Detects field types automatically
<!-- - **Validation** - Based on struct tags -->
<!-- - **Bulk Actions** - Mass edit/delete -->
<!-- - **Search & Filter** - Find data instantly -->

</td>
<td width="50%">

#### Developer Tools
- **API Playground** - Test endpoints live
- **Mock Mode** - Auto-rollback changes
- **Snapshots** - Save/restore database state
- **API Logger** - Monitor all requests
- **Mock Data** - Generate test data

</td>
</tr>
</table>

### Access

```bash
# Login page
http://localhost:8080/cms/login

# Default credentials
Username: superadmin
Password: superadmin
```

**Change credentials in:** `cms/main.go`

---

## Key Features

### Zero Configuration
No database, no config files, no setup hassles.

### Lightning Fast
Start coding business logic within minutes.

### Predictable State
Reset to initial data anytime or snapshot data anytime.
Perfect for testing.

### Extensible
Build on top of Gin framework. Easy to customize.

---


## Learn More

| Document | Description |
|----------|-------------|
| **[Documentation](./documentation.md)** | Core concepts, API reference, advanced usage |
---
## Roadmap

### Coming Soon
- [ ] **Project Structure Refactor** - separate CMS, CMS_Backend, Core
- [ ] **Query Parameters** - Filter, sort, search
- [ ] **Pagination** - Handle large datasets
- [ ] **Mass Delete** - Mass record delete
- [ ] **Model Generator** - Generate new model in CMS
- [ ] **Relationships** - Link between models
- [ ] **Export/Import** - JSON/CSV data exchange
- [ ] **Webhooks** - Event notifications
- [ ] **Scheduled Jobs** - Automated data / log / snapshot cleanup
- [ ] **Role-based Access** - Fine-grained permissions
- [ ] **File Uploads** - Multipart form support
- [ ] **Real Database Support** - exposing api for real dbs (Postgres, MySQL, etc)
---



**Migrate to a real database before going live.**

---

## Community & Support

- **Bug Reports:** [GitHub Issues](https://github.com/Minicode-HK/ezapi-go/issues)
- **Business Inquiries:** [Contact Us](mailto:hello@minicodehk.com)

---
