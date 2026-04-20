package api

import(
	"capstone/app/repositories"
	"github.com/gin-gonic/gin"
)

func GetAudit(r *gin.Context){
	audit, err := repositories.GetAudit()
	if err != nil{
		r.JSON(500, "Internal error")
		return
	}
	r.JSON(200, audit)
}