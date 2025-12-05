package app

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Minicode-HK/ezapi-go/ez"
)

type Product struct {
    Id          string  `json:"id"`
    Name        string  `json:"name" binding:"required"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Category    string  `json:"category"`

    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

var ProductDB []Product 


func init() {

    ez.New(&ProductDB).
        Seed([]Product{
            {Id: "1", Name: "Laptop", Price: 999.99, Category: "Electronics"},
            {Id: "2", Name: "Smartphone", Price: 499.99, Category: "Electronics"},
            {Id: "3", Name: "Desk Chair", Price: 89.99, Category: "Furniture"},
        }).
        CRUD("/api/products").
        CustomRoutes(func(router *gin.Engine) {
            // Get products by category
            router.GET("/api/products/category/:category", func(c *gin.Context) {
                filtered := ez.Query(ProductDB).Where("Category", "=", c.Param("category")).OrderBy("Price", false).Limit(2).Select("Category","Name").Get()
                ez.SendSuccess(c, filtered)
            })
        }).
        AfterCreate(func(newProduct *Product) error {
            log.Println("New product created:", newProduct.Name)
            return nil
        }).
        AutoTimestamp()
}