package core

import (
	"github.com/gin-gonic/gin"
)

func EnableCors(r *gin.Context) {
	r.Header("Access-Control-Allow-Origin", "http://localhost:4173")
	r.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	r.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

