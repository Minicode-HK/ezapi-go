package app

import (
	"github.com/gin-gonic/gin"

	"ezapi-go/ez"
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

    ez.New(&ProductDB).
        ResetableDB([]Product{
            {Id: "1", Name: "Laptop", Price: 999.99, Category: "Electronics"},
            {Id: "2", Name: "Smartphone", Price: 499.99, Category: "Electronics"},
            {Id: "3", Name: "Desk Chair", Price: 89.99, Category: "Furniture"},
        }).
        CRUDRoutes("/api/products").
        CustomRoutes(func(router *gin.Engine) {
        // Get products by category
        router.GET("/api/products/category/:category", func(c *gin.Context) {
            filtered := ez.Query(ProductDB).Where("Category", "==", c.Param("category")).OrderBy("Price", false).Limit(2).Get()
            ez.SendSuccess(c, filtered)
        })
    })

}