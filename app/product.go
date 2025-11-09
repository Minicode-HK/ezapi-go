package app

import (
    "strings"
    
    "github.com/gin-gonic/gin"
    
    "ezapi-go/core"
    "ezapi-go/core/feature"
    "ezapi-go/core/http"
)

type Product struct {
    Id          string  `json:"id"`
    Name        string  `json:"name" binding:"required"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Category    string  `json:"category"`
}

var ProductDB []Product 

func GetProductDB() *[]Product {
    return &ProductDB
}

func init() {

    // Initialize the in-memory database
    ProductDB = feature.ResetableDatabase(&ProductDB, []Product{
        {Id: "1", Name: "Laptop", Price: 999.99, Category: "Electronics"},
    })

    // Standard CRUD operations
    core.RegisterRouter(&ProductDB, "/api/products")
    
    // Custom endpoints
    core.RegisterRouterWith(func(router *gin.Engine) {
        // Get products by category
        router.GET("/api/products/category/:category", func(c *gin.Context) {
            category := c.Param("category")
            var filtered []Product
            
            for _, product := range ProductDB {
                if product.Category == category {
                    filtered = append(filtered, product)
                }
            }
            
            http.SendSuccess(c, filtered)
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
            
            http.SendSuccess(c, results)
        })
    })
}