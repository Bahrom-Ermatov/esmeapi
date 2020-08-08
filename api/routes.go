package api

// GetRoutes return routes for api
func (s *Server) getRoutes() {
	api := s.e.Group("/api/v1")
	api.GET("/balance/", s.GetClientBalance)
	api.GET("/sms/", s.GetSMSDetail)
	api.POST("/sms/send/", s.sendSMSView)
}
