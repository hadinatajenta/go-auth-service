package permission

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service} }
