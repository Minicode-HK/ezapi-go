package route

type Product struct {
    Id          string  `json:"id"`
    Name        string  `json:"name" binding:"required"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Category    string  `json:"category"`
}

var ProductDB []Product = []Product{
    {Id: "1", Name: "Laptop", Price: 999.99, Category: "Electronics"},
}

func init() {
    // Standard CRUD operations
    RegisterRouter(&ProductDB, "/api/products")
    
    // Custom endpoints
    RegisterRouterWith(func(router *gin.Engine) {
        // Get products by category
        router.GET("/api/products/category/:category", func(c *gin.Context) {
            category := c.Param("category")
            var filtered []Product
            
            for _, product := range ProductDB {
                if product.Category == category {
                    filtered = append(filtered, product)
                }
            }
            
            SendSuccess(c, filtered)
        })
        
        // Search products
        router.GET("/api/products/search", func(c *gin.Context) {
            query := c.Query("q")
            var results []Product
            
            for _, product := range ProductDB {
                if strings.Contains(strings.ToLower(product.Name), strings.ToLower(query)) {
                    results = append(results, product)
                }
            }
            
            SendSuccess(c, results)
        })
    })
}