package routes

import (
	"git.codenrock.com/avito/internal/handlers"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, tenderHandler *handlers.TenderHandler, bidHandler *handlers.BidHandler) {
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "ok",
			})
		})
		tenders := api.Group("/tenders")
		{
			tenders.GET("", tenderHandler.GetTenders)
			tenders.POST("/new", tenderHandler.CreateTender)
			tenders.GET("/my", tenderHandler.GetTendersByUsername)
			tenders.GET("/:tenderId/status", tenderHandler.GetTenderStatus)
			tenders.PUT("/:tenderId/status", tenderHandler.UpdateTenderStatus)
			tenders.PATCH("/tenders/:tenderId/edit", tenderHandler.UpdateTenderInfo)
			tenders.PATCH("/tenders/:tenderId/rollback/:version", tenderHandler.RollbackTenderVersion)
		}

		bids := api.Group("/bids")
		{
			bids.POST("/new", bidHandler.CreateBid)
			bids.GET("/my", bidHandler.GetUserBids)
			bids.GET("/:id/list", bidHandler.GetTenderBids)
			bids.GET("/:id/status", bidHandler.GetBidStatus)
			bids.PUT("/:id/status", bidHandler.UpdateBidStatus)
			bids.PATCH("/:id/edit", bidHandler.UpdateBid)
			bids.PUT("/:id/submit_decision", bidHandler.SubmitDecision)
			bids.PUT("/:id/feedback", bidHandler.SendFeedback)
			bids.PATCH("/:id/rollback/:version", bidHandler.RollbackBidVersion)
			bids.GET("/:id/reviews", bidHandler.GetBidReviews)
		}
	}
}
